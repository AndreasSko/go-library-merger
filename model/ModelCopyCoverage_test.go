package model

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

// Give every persisted field a nonzero value, including nested nullable fields,
// so omissions in the hand-written copy switch cannot hide behind zero values.
func populateCopyFields(value reflect.Value) {
	switch value.Kind() {
	case reflect.Pointer:
		value.Set(reflect.New(value.Type().Elem()))
		populateCopyFields(value.Elem())
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if value.Field(i).CanSet() {
				populateCopyFields(value.Field(i))
			}
		}
	case reflect.Slice:
		value.Set(reflect.MakeSlice(value.Type(), 2, 2))
		for i := 0; i < value.Len(); i++ {
			populateCopyFields(value.Index(i))
		}
	case reflect.String:
		value.SetString("copied value")
	case reflect.Bool:
		value.SetBool(true)
	case reflect.Int, reflect.Int32, reflect.Int64:
		value.SetInt(7)
	}
}

func TestMakeModelCopyPreservesAllFields(t *testing.T) {
	models := []Model{&BlockRange{}, &Bookmark{}, &InputField{}, &Location{}, &Note{}, &Tag{}, &TagMap{}, &UserMark{}, &IndependentMedia{}, &PlaylistItem{}, &PlaylistItemIndependentMediaMap{}, &PlaylistItemLocationMap{}, &UserMarkBlockRange{}}
	for _, original := range models {
		t.Run(fmt.Sprintf("%T", original), func(t *testing.T) {
			populateCopyFields(reflect.ValueOf(original).Elem())
			switch v := original.(type) {
			case *InputField:
				v.pseudoID = 11
			case *IndependentMedia:
				v.validatedHash, v.validatedSourceHash = "digest", v.Hash
			case *PlaylistItem:
				v.occurrence = 11
				v.tag = &Tag{TagID: 13, TagType: 2, Name: "P"}
				v.location = &Location{LocationID: 17, LocationType: 1}
				v.independentMedia = &IndependentMedia{IndependentMediaID: 19, Hash: "content", validatedHash: "digest", validatedSourceHash: "content"}
				v.thumbnail = &IndependentMedia{IndependentMediaID: 23, Hash: "thumbnail", validatedHash: "thumbnail-digest", validatedSourceHash: "thumbnail"}
			case *PlaylistItemIndependentMediaMap:
				v.pseudoID = 11
			case *PlaylistItemLocationMap:
				v.pseudoID = 11
			}
			copied := MakeModelCopy(original)
			require.Equal(t, original, copied)
			require.NotSame(t, original, copied)
			if item, ok := original.(*PlaylistItem); ok {
				cp := copied.(*PlaylistItem)
				require.NotSame(t, item.tag, cp.tag)
				require.NotSame(t, item.location, cp.location)
				require.NotSame(t, item.independentMedia, cp.independentMedia)
				require.NotSame(t, item.thumbnail, cp.thumbnail)
			}
		})
	}
}
