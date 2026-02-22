package merger

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AndreasSko/go-library-merger/model"
)

// MergeIndependentMedia tries to merge the left and right slice of IndependentMedia. If there is a
// collision, it returns an error asking for specification how it should handle it.
func MergeIndependentMedia(left []*model.IndependentMedia, right []*model.IndependentMedia) ([]*model.IndependentMedia, IDChanges, error) {
	result, changes, err := tryMergeWithConflictSolver(left, right, nil, solveIndependentMediaMergeConflict)

	return model.IndependentMedia{}.MakeSlice(result), changes, err
}

// remapIndependentMediaPaths gives distinct media that use the same archive
// path separate deterministic destinations. The lowest content hash keeps the
// original path, making the result independent of input order.
func remapIndependentMediaPaths(media []*model.IndependentMedia) (map[string]string, error) {
	byPath := map[string][]*model.IndependentMedia{}
	reserved := map[string]bool{"userdata.db": true, "manifest.json": true, "default_thumbnail.png": true}
	for _, im := range media {
		if im == nil {
			continue
		}
		if err := validateMediaPath(im.FilePath); err != nil {
			return nil, err
		}
		key := strings.ToLower(im.FilePath)
		byPath[key] = append(byPath[key], im)
		reserved[key] = true
	}

	remapped := map[string]string{}
	for _, group := range byPath {
		sort.Slice(group, func(i, j int) bool { return group[i].UniqueKey() < group[j].UniqueKey() })
		for i, im := range group {
			path := im.FilePath
			key := im.UniqueKey() + "\x00" + path
			archiveReserved := strings.EqualFold(path, "userData.db") || strings.EqualFold(path, "manifest.json") || strings.EqualFold(path, "default_thumbnail.png")
			if i == 0 && !archiveReserved {
				remapped[key] = path
				continue
			}
			digest := sha256.Sum256([]byte(im.UniqueKey()))
			digestHex := fmt.Sprintf("%x", digest)
			ext := filepath.Ext(path)
			stem := strings.TrimSuffix(path, ext)
			for length := 12; ; length += 4 {
				if length > len(digestHex) {
					length = len(digestHex)
				}
				candidate := stem + "-" + digestHex[:length] + ext
				if !reserved[strings.ToLower(candidate)] {
					im.FilePath = candidate
					remapped[key] = candidate
					reserved[strings.ToLower(candidate)] = true
					break
				}
				if length == len(digestHex) {
					return nil, fmt.Errorf("could not allocate media path for hash %s", im.Hash)
				}
			}
		}
	}
	return remapped, nil
}

// CopyMergedIndependentMedia copies from the input that owns the selected
// (hash, path) pair. A path alone is insufficient: another backup can use the
// same filename for different content.
func CopyMergedIndependentMedia(ims []*model.IndependentMedia, left, right *model.Database, dst string) error {
	type source struct {
		dir  string
		path string
	}
	sources := map[string][]source{}
	for _, db := range []*model.Database{left, right} {
		for _, im := range db.IndependentMedia {
			if im != nil {
				sources[im.UniqueKey()] = append(sources[im.UniqueKey()], source{db.IndependentMediaDir, im.FilePath})
			}
		}
	}
	for _, im := range ims {
		if im == nil {
			continue
		}
		copied := false
		candidates := sources[im.UniqueKey()]
		sort.SliceStable(candidates, func(i, j int) bool {
			return candidates[i].path == im.FilePath && candidates[j].path != im.FilePath
		})
		for _, source := range candidates {
			if err := copyMediaFile(source.dir, source.path, dst, im.FilePath); err == nil {
				copied = true
				break
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		if !copied {
			return fmt.Errorf("missing source media %q (%s)", im.FilePath, im.Hash)
		}
	}
	return nil
}

func validateMediaPath(path string) error {
	if path == "." || path == ".." || !filepath.IsLocal(path) || filepath.Base(path) != path || strings.Contains(path, "\\") {
		return fmt.Errorf("invalid independent media path %q", path)
	}
	return nil
}

func copyMediaFile(srcDir, srcPath, dstDir, dstPath string) (resultErr error) {
	if err := validateMediaPath(srcPath); err != nil {
		return err
	}
	if err := validateMediaPath(dstPath); err != nil {
		return err
	}
	src, err := os.Open(filepath.Join(srcDir, srcPath))
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, src.Close()) }()
	dst, err := os.Create(filepath.Join(dstDir, dstPath))
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, dst.Close()) }()
	_, err = io.Copy(dst, src)
	return err
}

// solveIndependentMediaMergeConflict solves a merge conflict for IndependentMedia by choosing the side that more likely
// contains the original file name. A "." in the filename is considered an indicator that the name is the original filename.
func solveIndependentMediaMergeConflict(conflicts map[string]MergeConflict) (map[string]MergeSolution, error) {
	solution := make(map[string]MergeSolution, len(conflicts))

	for key, value := range conflicts {
		l, ok := value.Left.(*model.IndependentMedia)
		if !ok {
			return nil, fmt.Errorf("expected *model.IndependentMedia, got %T", value.Left)
		}
		r, ok := value.Right.(*model.IndependentMedia)
		if !ok {
			return nil, fmt.Errorf("expected *model.IndependentMedia, got %T", value.Right)
		}

		if l.MimeType != r.MimeType {
			return nil, fmt.Errorf("unexpected mime type mismatch for %v and %v", l, r)
		}

		if strings.Contains(l.OriginalFilename, ".") {
			solution[key] = MergeSolution{Side: LeftSide, Solution: value.Left, Discarded: value.Right}
		} else {
			solution[key] = MergeSolution{Side: RightSide, Solution: value.Right, Discarded: value.Left}
		}
	}

	return solution, nil
}
