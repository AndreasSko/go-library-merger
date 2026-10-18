package publication

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/sync/singleflight"

	"github.com/AndreasSko/go-library-merger/model"
)

// MediaLinksURL is the URL to the jw.org API that lists the download links of
// a publication
var MediaLinksURL = "https://b.jw-cdn.org/apis/pub-media/GETPUBMEDIALINKS"

// client limits the time of a request, so a stalled download cannot block
// a conflict view
var client = &http.Client{Timeout: 2 * time.Minute}

// Snippet is the text of a paragraph, split around the tokens of a marking
type Snippet struct {
	Before string `json:"before"`
	Marked string `json:"marked"`
	After  string `json:"after"`
	// Number is the paragraph number shown in the publication, or 0 if it
	// has none
	Number int `json:"number,omitempty"`
	// Heading is the subheading the paragraph belongs to
	Heading string `json:"heading,omitempty"`
}

// paragraph is the text of a paragraph and where it is in its document
type paragraph struct {
	Text    string `json:"text"`
	Number  int    `json:"number,omitempty"`
	Heading string `json:"heading,omitempty"`
}

// documents maps a document ID to its paragraphs
type documents map[int64]map[int]paragraph

var keySymbolPattern = regexp.MustCompile(`^[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$`)

// A token is a word or a single other character, like a punctuation mark. A
// single character between two words keeps them together, so "36:17" and
// "Jehovah’s" are one token each. Spaces separate tokens, including no-break
// and zero-width ones.
var token = func() *regexp.Regexp {
	word := `[\p{L}\p{N}\p{M}]+`
	other := `[^\p{L}\p{N}\p{M}\s\p{Z}\p{Cf}]`
	return regexp.MustCompile(word + `(?:` + other + word + `)*|` + other)
}()

// MarkingSnippets returns the text of the paragraphs that a marking covers.
// Paragraphs that are not available are left out. The text comes from the
// EPUB of the publication on jw.org. Only the paragraph text is kept in
// cacheDir, so each publication is downloaded once.
func MarkingSnippets(ctx context.Context, cacheDir string, location *model.Location, blockRanges []*model.BlockRange) ([]Snippet, error) {
	path, ok := cachePath(cacheDir, location, blockRanges)
	if !ok {
		return nil, nil
	}
	docs, err, _ := lookups.Do(path, func() (any, error) {
		return lookupDocuments(ctx, path, location)
	})
	if err != nil {
		return nil, err
	}

	var snippets []Snippet
	for _, br := range blockRanges {
		if br.BlockType != 1 || !br.StartToken.Valid || !br.EndToken.Valid {
			continue
		}
		p := docs.(documents)[location.DocumentID.Int64][br.Identifier]
		if snippet, ok := markTokens(p.Text, int(br.StartToken.Int32), int(br.EndToken.Int32)); ok {
			snippet.Number, snippet.Heading = p.Number, p.Heading
			snippets = append(snippets, snippet)
		}
	}
	return snippets, nil
}

// NeedsDownload reports whether MarkingSnippets has to download the
// publication of a marking first, because its text is not in cacheDir yet.
func NeedsDownload(cacheDir string, location *model.Location, blockRanges []*model.BlockRange) bool {
	path, ok := cachePath(cacheDir, location, blockRanges)
	if !ok {
		return false
	}
	_, err := os.Stat(path)
	return err != nil
}

// cachePath returns the file in cacheDir with the text of the publication of
// a marking. It returns false if the marking has no paragraph of a
// publication that can be looked up.
func cachePath(cacheDir string, location *model.Location, blockRanges []*model.BlockRange) (string, bool) {
	// Block type 1 is a paragraph of a publication, unlike a Bible verse
	hasParagraph := slices.ContainsFunc(blockRanges, func(br *model.BlockRange) bool { return br.BlockType == 1 })
	if !hasParagraph || location == nil || !location.DocumentID.Valid || !location.MepsLanguage.Valid {
		return "", false
	}
	lang := model.MepsLanguageSymbol(int(location.MepsLanguage.Int32))
	if lang == "" || !keySymbolPattern.MatchString(location.KeySymbol.String) {
		return "", false
	}
	return filepath.Join(cacheDir, fmt.Sprintf("%s_%s_%d.json", location.KeySymbol.String, lang, location.IssueTagNumber)), true
}

// lookups makes concurrent lookups of the same publication, like for both
// sides of a conflict, share one download
var lookups singleflight.Group

// lookupDocuments returns the documents of a publication from the cache file
// at path, or downloads them first
func lookupDocuments(ctx context.Context, path string, location *model.Location) (documents, error) {
	docs := documents{}
	// A cache file that cannot be read is replaced by a new download
	if data, err := os.ReadFile(path); err == nil && json.Unmarshal(data, &docs) == nil {
		return docs, nil
	}

	cacheDir := filepath.Dir(path)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	epub, err := downloadEPUB(ctx, cacheDir, location)
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.Remove(epub) }()
	if docs, err = readEPUB(epub); err != nil {
		return nil, err
	}

	data, err := json.Marshal(docs)
	if err != nil {
		return nil, err
	}
	tmp, err := writeTemp(cacheDir, "*.json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to cache publication: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return nil, fmt.Errorf("failed to cache publication: %w", err)
	}
	return docs, nil
}

// downloadEPUB downloads the EPUB of the publication of location into dir and
// returns its path
func downloadEPUB(ctx context.Context, dir string, location *model.Location) (string, error) {
	keySymbol := location.KeySymbol.String
	lang := model.MepsLanguageSymbol(int(location.MepsLanguage.Int32))
	issueTagNumber := location.IssueTagNumber
	query := url.Values{
		"pub":         {keySymbol},
		"langwritten": {lang},
		"fileformat":  {"EPUB"},
		"output":      {"json"},
	}
	// The API leaves out the day of monthly issues, like 202004 for 20200400
	if issueTagNumber%100 == 0 {
		issueTagNumber /= 100
	}
	if issueTagNumber != 0 {
		query.Set("issue", strconv.Itoa(issueTagNumber))
	}
	var links struct {
		Files map[string]struct {
			EPUB []struct {
				File struct {
					URL string `json:"url"`
				} `json:"file"`
			} `json:"EPUB"`
		} `json:"files"`
	}
	body, err := get(ctx, MediaLinksURL+"?"+query.Encode())
	if err != nil {
		return "", err
	}
	err = json.NewDecoder(body).Decode(&links)
	if err = errors.Join(err, body.Close()); err != nil {
		return "", fmt.Errorf("failed to read links of publication %s: %w", keySymbol, err)
	}
	files := links.Files[lang].EPUB
	if len(files) == 0 {
		return "", fmt.Errorf("no EPUB found for publication %s", keySymbol)
	}

	epub, err := get(ctx, files[0].File.URL)
	if err != nil {
		return "", err
	}
	path, err := writeTemp(dir, "*.epub", epub)
	if err = errors.Join(err, epub.Close()); err != nil {
		if path != "" {
			_ = os.Remove(path)
		}
		return "", fmt.Errorf("failed to download EPUB: %w", err)
	}
	return path, nil
}

// writeTemp writes r to a new temporary file in dir and returns its path
func writeTemp(dir string, pattern string, r io.Reader) (string, error) {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(f, r)
	if err = errors.Join(err, f.Close()); err != nil {
		_ = os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

func get(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request %s: %w", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Join(fmt.Errorf("request to %s failed with status %s", url, resp.Status), resp.Body.Close())
	}
	return resp.Body, nil
}

var documentFile = regexp.MustCompile(`^OEBPS/(\d+)\.xhtml$`)

// readEPUB reads the paragraphs of all documents in an EPUB. Each document is
// a file named after its MEPS document ID.
func readEPUB(path string) (docs documents, resultErr error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, r.Close()) }()

	docs = documents{}
	for _, f := range r.File {
		match := documentFile.FindStringSubmatch(f.Name)
		if match == nil {
			continue
		}
		id, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return nil, err
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %w", f.Name, err)
		}
		docs[id], err = readParagraphs(rc)
		if err = errors.Join(err, rc.Close()); err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", f.Name, err)
		}
	}
	return docs, nil
}

// readParagraphs reads the text of all elements with a data-pid, which is
// the paragraph ID JW Library uses. Paragraph numbers and footnote markers
// are left out of the text, as JW Library does not count them as tokens.
func readParagraphs(r io.Reader) (map[int]paragraph, error) {
	paragraphs := map[int]paragraph{}
	z := html.NewTokenizer(r)
	pid := -1
	var text strings.Builder
	var current paragraph
	isHeading := false
	heading := ""
	var skipped []bool // per open element inside the paragraph
	skipping := 0
	for {
		switch z.Next() {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				return paragraphs, nil
			}
			return nil, z.Err()
		case html.StartTagToken:
			tag := z.Token()
			if pid < 0 {
				id, err := strconv.Atoi(attr(tag, "data-pid"))
				if err != nil {
					continue
				}
				pid = id
				text.Reset()
				// h1 is the title of the document, the others are subheadings
				isHeading = slices.Contains([]string{"h2", "h3", "h4", "h5", "h6"}, tag.Data)
				current = paragraph{}
				if !isHeading {
					current.Heading = heading
				}
			}
			if number, err := strconv.Atoi(attr(tag, "data-pnum")); err == nil {
				current.Number = number
			}
			skip := tag.Data == "sup" || tag.Data == "a" && strings.HasPrefix(attr(tag, "href"), "#footnote")
			if skip {
				skipping++
			}
			skipped = append(skipped, skip)
		case html.EndTagToken:
			if pid < 0 {
				continue
			}
			if skipped[len(skipped)-1] {
				skipping--
			}
			skipped = skipped[:len(skipped)-1]
			if len(skipped) == 0 {
				// Drop soft hyphens, so they do not split words into tokens
				current.Text = strings.TrimSpace(strings.ReplaceAll(text.String(), "\u00ad", ""))
				paragraphs[pid] = current
				if isHeading {
					heading = current.Text
				}
				pid = -1
			}
		case html.TextToken:
			if pid >= 0 && skipping == 0 {
				text.Write(z.Text())
			}
		}
	}
}

func attr(t html.Token, key string) string {
	for _, a := range t.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// markTokens splits text around the tokens from start to end. Like in JW
// Library, the first token has the index 0. It returns false if the tokens
// are not part of the text.
func markTokens(text string, start int, end int) (Snippet, bool) {
	tokens := token.FindAllStringIndex(text, -1)
	if start < 0 || start > end || end >= len(tokens) {
		return Snippet{}, false
	}
	from, to := tokens[start][0], tokens[end][1]
	return Snippet{Before: text[:from], Marked: text[from:to], After: text[to:]}, true
}

// maxContext is the number of characters that are kept at most before and
// after the marked text
const maxContext = 80

// sentenceEnd matches the end of a sentence and the space after it. It needs
// a word of at least two letters or a closing bracket before the punctuation
// mark, so abbreviations like "C. T. Russell" or "z. B." do not end one, and
// a capital letter or an opening quote after it, so references like
// "Joh. 4:3" do not either.
var sentenceEnd = regexp.MustCompile(`(?:\p{L}{2}|\))[.!?]+[”“"’'»)\]]*(\s+)[\p{Lu}„“"‘«(]`)

// Shorten keeps only the sentences of a snippet that contain the marked text.
// Context that is still longer than maxContext is cut at a word and marked
// with "…".
func Shorten(s Snippet) Snippet {
	text := s.Before + s.Marked + s.After
	from, to := len(s.Before), len(s.Before)+len(s.Marked)
	start, end := 0, len(text)
	for _, m := range sentenceEnd.FindAllStringSubmatchIndex(text, -1) {
		spaceStart, spaceEnd := m[2], m[3]
		if spaceEnd <= from {
			start = spaceEnd
		}
		if spaceStart >= to && end == len(text) {
			end = spaceStart
		}
	}

	before, after := text[start:from], text[to:end]
	if r := []rune(before); len(r) > maxContext {
		before = string(r[len(r)-maxContext:])
		before = "…" + before[strings.IndexByte(before, ' ')+1:]
	}
	if r := []rune(after); len(r) > maxContext {
		after = string(r[:maxContext])
		if i := strings.LastIndexByte(after, ' '); i > 0 {
			after = after[:i]
		}
		after = strings.TrimRight(after, ",;:–— ") + "…"
	}
	s.Before, s.After = before, after
	return s
}
