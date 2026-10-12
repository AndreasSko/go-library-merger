package model

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
)

var playlistSourceSymbol = regexp.MustCompile(`^[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$`)

// PlaylistSourceURL returns an official source link when the imported identity
// is sufficient to identify it. It never uses the editable playlist label or
// guesses a language, multimedia type, or a dated media identifier.
func PlaylistSourceURL(location *Location, majorMultimediaType int) string {
	if location == nil || !location.MepsLanguage.Valid {
		return ""
	}
	locale := map[int32]string{0: "E", 2: "X"}[location.MepsLanguage.Int32]
	if locale == "" {
		return ""
	}
	query := url.Values{"wtlocale": {locale}}
	if location.LocationType == 3 {
		if majorMultimediaType != 2 || !location.KeySymbol.Valid ||
			!playlistSourceSymbol.MatchString(location.KeySymbol.String) ||
			!location.Track.Valid || location.Track.Int32 <= 0 || location.IssueTagNumber != 0 ||
			(location.Specialty.Valid && location.Specialty.String != "") ||
			(location.Edition.Valid && location.Edition.String != "") {
			return ""
		}
		query.Set("lank", fmt.Sprintf("pub-%s_%d_VIDEO", location.KeySymbol.String, location.Track.Int32))
	} else if location.DocumentID.Valid && location.DocumentID.Int64 > 0 {
		query.Set("docid", strconv.FormatInt(location.DocumentID.Int64, 10))
	} else {
		return ""
	}
	return "https://www.jw.org/finder?" + query.Encode()
}
