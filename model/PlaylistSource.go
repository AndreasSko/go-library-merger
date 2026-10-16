package model

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
)

var playlistSourceSymbol = regexp.MustCompile(`^[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$`)

// PlaylistSourceURL returns a jw.org link to the location, or "" when the
// location does not identify the source clearly. The playlist label is not
// used, because the user can change it.
func PlaylistSourceURL(location *Location, majorMultimediaType int) string {
	if location == nil || !location.MepsLanguage.Valid {
		return ""
	}
	locale := mepsLanguageSymbols[location.MepsLanguage.Int32]
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
