package publication

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AndreasSko/go-library-merger/model"
)

func TestMarkingSnippets(t *testing.T) {
	var epub bytes.Buffer
	zw := zip.NewWriter(&epub)
	f, err := zw.Create("OEBPS/2020360.xhtml")
	require.NoError(t, err)
	_, err = f.Write([]byte(`<html><body>
<h1 id="p1" data-pid="1">Title</h1>
<h2 id="p4" data-pid="4">A Subheading</h2>
<p id="p2" data-pid="2"><span class="parNum" data-pnum="1"><strong><sup>1</sup></strong></span> Jehovah’s people read Joel&#160;2:7. They lis&#173;tened.<a href="#footnote1">*</a></p>
</body></html>`))
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/links":
			assert.Equal(t, "w", r.URL.Query().Get("pub"))
			assert.Equal(t, "202004", r.URL.Query().Get("issue"))
			assert.Equal(t, "X", r.URL.Query().Get("langwritten"))
			_, err := fmt.Fprintf(w, `{"files":{"X":{"EPUB":[{"file":{"url":"%s/w_X_202004.epub"}}]}}}`, server.URL)
			assert.NoError(t, err)
		case "/w_X_202004.epub":
			_, err := w.Write(epub.Bytes())
			assert.NoError(t, err)
		}
	}))
	oldURL := MediaLinksURL
	t.Cleanup(func() { MediaLinksURL = oldURL })
	MediaLinksURL = server.URL + "/links"

	location := &model.Location{
		DocumentID:     sql.NullInt64{Int64: 2020360, Valid: true},
		IssueTagNumber: 20200400,
		KeySymbol:      sql.NullString{String: "w", Valid: true},
		MepsLanguage:   sql.NullInt32{Int32: 2, Valid: true},
	}
	blockRanges := []*model.BlockRange{
		{BlockType: 1, Identifier: 2, StartToken: sql.NullInt32{Int32: 3, Valid: true}, EndToken: sql.NullInt32{Int32: 5, Valid: true}},
		{BlockType: 1, Identifier: 3, StartToken: sql.NullInt32{Int32: 0, Valid: true}, EndToken: sql.NullInt32{Int32: 1, Valid: true}},
		{BlockType: 2, Identifier: 2, StartToken: sql.NullInt32{Int32: 0, Valid: true}, EndToken: sql.NullInt32{Int32: 1, Valid: true}},
	}
	expected := []Snippet{{
		Before:  "Jehovah’s people read ",
		Marked:  "Joel\u00a02:7.",
		After:   " They listened.",
		Number:  1,
		Heading: "A Subheading",
	}}

	cacheDir := t.TempDir()
	assert.True(t, NeedsDownload(cacheDir, location, blockRanges))
	assert.False(t, NeedsDownload(cacheDir, location, blockRanges[2:]), "Bible verses are not looked up")
	snippets, err := MarkingSnippets(context.Background(), cacheDir, location, blockRanges)
	require.NoError(t, err)
	assert.Equal(t, expected, snippets)
	assert.False(t, NeedsDownload(cacheDir, location, blockRanges))

	server.Close()
	snippets, err = MarkingSnippets(context.Background(), cacheDir, location, blockRanges)
	require.NoError(t, err)
	assert.Equal(t, expected, snippets, "should be read from the cache")
}

func Test_markTokens(t *testing.T) {
	text := "In 33\u00a0C.E., Jehovah’s “people”\u200b—Acts 2:41."
	tests := []struct {
		start, end int
		marked     string
		ok         bool
	}{
		{0, 0, "In", true},
		{1, 4, "33\u00a0C.E.,", true},
		{5, 8, "Jehovah’s “people”", true},
		{9, 12, "—Acts 2:41.", true},
		{9, 13, "", false},
		{2, 1, "", false},
	}
	for _, tt := range tests {
		snippet, ok := markTokens(text, tt.start, tt.end)
		assert.Equal(t, tt.ok, ok)
		assert.Equal(t, tt.marked, snippet.Marked)
	}
}

func TestShorten(t *testing.T) {
	long := strings.Repeat("word ", 30)
	tests := []struct {
		name     string
		input    Snippet
		expected Snippet
	}{
		{
			name:     "keeps the sentence of the marked text",
			input:    Snippet{Before: "First one. Bruder C. T. Russell read ", Marked: "Joh. 4:3", After: " (Joh. 4:3-6). Then he left."},
			expected: Snippet{Before: "Bruder C. T. Russell read ", Marked: "Joh. 4:3", After: " (Joh. 4:3-6)."},
		},
		{
			name:     "cuts long context at a word",
			input:    Snippet{Before: long, Marked: "marked", After: ", " + long + "end."},
			expected: Snippet{Before: "…" + strings.TrimSpace(strings.Repeat("word ", 15)) + " ", Marked: "marked", After: ", " + strings.TrimSpace(strings.Repeat("word ", 15)) + "…"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Shorten(tt.input))
		})
	}
}
