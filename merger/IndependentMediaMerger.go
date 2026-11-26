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

// MergeIndependentMedia merges the left and right slice of IndependentMedia.
// Media with the same content is kept once. If its name or path differ, the
// left one is kept, since the content is the same.
func MergeIndependentMedia(left []*model.IndependentMedia, right []*model.IndependentMedia) ([]*model.IndependentMedia, IDChanges, error) {
	chooseLeft := func(conflicts map[string]MergeConflict) (map[string]MergeSolution, error) {
		return AutoResolveConflicts(conflicts, "chooseLeft")
	}
	result, changes, err := tryMergeWithConflictSolver(left, right, nil, chooseLeft)

	return model.IndependentMedia{}.MakeSlice(result), changes, err
}

// dedupeIndependentMedia keeps one row for media that a backup has several
// times with the same content. The merge keeps only one row per content
// within a backup too, but it would move only that row's references, so
// playlist items of the other rows would point to the wrong media. The rows
// of the other items are moved to the kept row here, which prefers a row
// whose file is included. Only rows with the same key are merged, so the
// keys and repeat indexes of the playlist items stay as they are.
func dedupeIndependentMedia(db *model.Database) {
	kept := map[string]*model.IndependentMedia{}
	for _, im := range db.IndependentMedia {
		if im == nil {
			continue
		}
		if k, ok := kept[im.UniqueKey()]; !ok || (k.FileMissing() && !im.FileMissing()) {
			kept[im.UniqueKey()] = im
		}
	}
	ids, paths := map[int]int{}, map[string]string{}
	for i, im := range db.IndependentMedia {
		if im == nil || kept[im.UniqueKey()] == im {
			continue
		}
		ids[im.IndependentMediaID] = kept[im.UniqueKey()].IndependentMediaID
		paths[im.FilePath] = kept[im.UniqueKey()].FilePath
		db.IndependentMedia[i] = nil
	}
	if len(ids) == 0 {
		return
	}
	for _, m := range db.PlaylistItemIndependentMediaMap {
		if m == nil {
			continue
		}
		if id, ok := ids[m.IndependentMediaID]; ok {
			m.IndependentMediaID = id
		}
	}
	for _, item := range db.PlaylistItem {
		if item != nil && item.ThumbnailFilePath.Valid {
			if path, ok := paths[item.ThumbnailFilePath.String]; ok {
				item.ThumbnailFilePath.String = path
			}
		}
	}
}

// remapIndependentMediaPaths renames media whose filename is already used by
// other media, for example IMG_0001.jpg from two different phones. The media
// with the lowest hash keeps its name, so the result does not depend on the
// order of the backups. The others get a part of their hash added.
func remapIndependentMediaPaths(media []*model.IndependentMedia) (map[string]string, error) {
	// Backup files are compared case-insensitively, as on macOS and Windows.
	used := map[string]bool{"userdata.db": true, "manifest.json": true, "default_thumbnail.png": true}
	byPath := map[string][]*model.IndependentMedia{}
	for _, im := range media {
		if im == nil {
			continue
		}
		byPath[strings.ToLower(im.FilePath)] = append(byPath[strings.ToLower(im.FilePath)], im)
	}
	paths := make([]string, 0, len(byPath))
	for path := range byPath {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	remapped := map[string]string{}
	var renamed []*model.IndependentMedia
	for _, path := range paths {
		group := byPath[path]
		sort.Slice(group, func(i, j int) bool { return group[i].UniqueKey() < group[j].UniqueKey() })
		if used[path] {
			renamed = append(renamed, group...)
		} else {
			used[path] = true
			remapped[group[0].UniqueKey()+"\x00"+group[0].FilePath] = group[0].FilePath
			renamed = append(renamed, group[1:]...)
		}
	}
	for _, im := range renamed {
		// The hash comes from the backup, so hash it again to get a safe filename.
		suffix := fmt.Sprintf("%x", sha256.Sum256([]byte(im.UniqueKey())))[:12]
		ext := filepath.Ext(im.FilePath)
		path := strings.TrimSuffix(im.FilePath, ext) + "-" + suffix + ext
		if used[strings.ToLower(path)] {
			return nil, fmt.Errorf("media path %q is already used", path)
		}
		used[strings.ToLower(path)] = true
		remapped[im.UniqueKey()+"\x00"+im.FilePath] = path
		im.FilePath = path
	}
	return remapped, nil
}

// CopyMergedIndependentMedia copies the files of the merged media from the
// backups into dst. Files are looked up by their content, because another
// backup can use the same filename for different content. Media whose file
// is missing in both backups is left without a file.
func CopyMergedIndependentMedia(ims []*model.IndependentMedia, left, right *model.Database, dst string) error {
	sources := map[string]string{}
	for _, db := range []*model.Database{left, right} {
		for _, im := range db.IndependentMedia {
			if im != nil && !im.FileMissing() {
				sources[im.UniqueKey()] = filepath.Join(db.WorkingDir, im.FilePath)
			}
		}
	}
	for _, im := range ims {
		if im == nil {
			continue
		}
		source, ok := sources[im.UniqueKey()]
		if !ok && im.FileMissing() {
			continue
		}
		if !ok {
			return fmt.Errorf("missing source media %q", im.FilePath)
		}
		if err := copyFile(source, filepath.Join(dst, im.FilePath)); err != nil {
			return err
		}
		im.SetFileMissing(false)
	}
	return nil
}

func copyFile(srcPath, dstPath string) (resultErr error) {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, src.Close()) }()
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, dst.Close()) }()
	_, err = io.Copy(dst, src)
	return err
}
