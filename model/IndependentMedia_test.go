package model

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchMissingMedia(t *testing.T) {
	canonical := "0a11223344556677889900aabbccddeeff11223344556677889900aabbccddee"
	digest, err := hex.DecodeString(canonical)
	assert.NoError(t, err)
	native := nativeHash(digest)
	assert.NotEqual(t, canonical, native)
	missing := func(hash string) *IndependentMedia {
		return &IndependentMedia{Hash: hash, fileMissing: true}
	}

	left := &Database{IndependentMedia: []*IndependentMedia{nil, missing(native), missing(native)}}
	right := &Database{IndependentMedia: []*IndependentMedia{nil, {Hash: canonical, contentHash: canonical}}}
	MatchMissingMedia(left, right)
	assert.Equal(t, canonical, left.IndependentMedia[1].UniqueKey())
	assert.Equal(t, canonical, left.IndependentMedia[2].UniqueKey())

	// The backup has the file under another row already, so the missing row
	// keeps its own key instead of becoming a duplicate of that row.
	both := &Database{IndependentMedia: []*IndependentMedia{nil, {Hash: canonical, contentHash: canonical}, missing(native)}}
	MatchMissingMedia(both, right)
	assert.Equal(t, native, both.IndependentMedia[2].UniqueKey())

	// The same with the hash in upper case.
	upper := &Database{IndependentMedia: []*IndependentMedia{nil, {Hash: canonical, contentHash: canonical}, missing(strings.ToUpper(canonical))}}
	MatchMissingMedia(upper, right)
	assert.Equal(t, strings.ToUpper(canonical), upper.IndependentMedia[2].UniqueKey())

	// Two missing rows of one backup with different forms of the same hash
	// can't both get the key.
	forms := &Database{IndependentMedia: []*IndependentMedia{nil, missing(native), missing(strings.ToUpper(canonical))}}
	MatchMissingMedia(forms, right)
	assert.NotEqual(t, strings.ToLower(forms.IndependentMedia[1].UniqueKey()), strings.ToLower(forms.IndependentMedia[2].UniqueKey()))

	// Hashes that differ only in case had different keys, and keep them.
	cases := &Database{IndependentMedia: []*IndependentMedia{nil, missing(canonical), missing(strings.ToUpper(canonical))}}
	MatchMissingMedia(cases, right)
	assert.Equal(t, canonical, cases.IndependentMedia[1].UniqueKey())
	assert.Equal(t, strings.ToUpper(canonical), cases.IndependentMedia[2].UniqueKey())

	// Two groups match the same key: the first in sort order gets it, every
	// time.
	for range 20 {
		lower := &Database{IndependentMedia: []*IndependentMedia{nil, missing(native), missing(strings.ToUpper(native))}}
		MatchMissingMedia(lower, right)
		assert.Equal(t, native, lower.IndependentMedia[1].UniqueKey())
		assert.Equal(t, canonical, lower.IndependentMedia[2].UniqueKey())
	}

	// An upper case hash without another row gets the key.
	alone := &Database{IndependentMedia: []*IndependentMedia{nil, missing(strings.ToUpper(canonical))}}
	MatchMissingMedia(alone, right)
	assert.Equal(t, canonical, alone.IndependentMedia[1].UniqueKey())

	// Unknown media stays as it is.
	other := &Database{IndependentMedia: []*IndependentMedia{nil, missing("ff")}}
	MatchMissingMedia(other, right)
	assert.Equal(t, "ff", other.IndependentMedia[1].UniqueKey())
}
