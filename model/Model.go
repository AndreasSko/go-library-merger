package model

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/mitchellh/go-wordwrap"
)

// Model represents a general table of the JW Library database and
// defines methods that are needed so entries are mergeable.
type Model interface {
	ID() int
	SetID(int)
	UniqueKey() string
	Equals(m2 Model) bool
	RelatedEntries(db *Database) Related
	PrettyPrint(db *Database) string
	tableName() string
	idName() string
	scanRow(row *sql.Rows) (Model, error)
}

// Related combines entries that are related to a given model
type Related struct {
	BlockRange          []*BlockRange       `json:"blockRange"`
	Bookmark            *Bookmark           `json:"bookmark"`
	Location            *Location           `json:"location"`
	PublicationLocation *Location           `json:"publicationLocation"`
	Note                *Note               `json:"note"`
	Tag                 *Tag                `json:"tag"`
	TagMap              *TagMap             `json:"tagMap"`
	UserMark            *UserMark           `json:"userMark"`
	UserMarkBlockRange  *UserMarkBlockRange `json:"userMarkBlockRange"`
}

// MakeModelSlice converts a slice of pointers of model-implementing structs to []model
func MakeModelSlice(arg interface{}) ([]Model, error) {
	slice := reflect.ValueOf(arg)

	if slice.Kind() != reflect.Kind(reflect.Slice) {
		return nil, fmt.Errorf("Can't create []model out of %T", arg)
	}

	c := slice.Len()
	result := make([]Model, c)
	for i := 0; i < c; i++ {
		result[i] = slice.Index(i).Interface().(Model)
	}
	return result, nil
}

// MakeModelCopy copies the content of the given Model (pointer to a
// model-implementing struct) to a new Model
func MakeModelCopy(mdl Model) Model {
	var mdlCopy Model
	switch mdl.(type) {
	case *BlockRange:
		mdl := mdl.(*BlockRange)
		mdlCopy = &BlockRange{
			BlockRangeID: mdl.BlockRangeID,
			BlockType:    mdl.BlockType,
			Identifier:   mdl.Identifier,
			StartToken:   sql.NullInt32{Int32: mdl.StartToken.Int32, Valid: mdl.StartToken.Valid},
			EndToken:     sql.NullInt32{Int32: mdl.EndToken.Int32, Valid: mdl.EndToken.Valid},
			UserMarkID:   mdl.UserMarkID,
		}
	case *Bookmark:
		mdl := mdl.(*Bookmark)
		mdlCopy = &Bookmark{
			BookmarkID:            mdl.BookmarkID,
			LocationID:            mdl.LocationID,
			PublicationLocationID: mdl.PublicationLocationID,
			Slot:                  mdl.Slot,
			Title:                 mdl.Title,
			Snippet:               sql.NullString{String: mdl.Snippet.String, Valid: mdl.Snippet.Valid},
			BlockType:             mdl.BlockType,
			BlockIdentifier:       sql.NullInt32{Int32: mdl.BlockIdentifier.Int32, Valid: mdl.BlockIdentifier.Valid},
		}
	case *InputField:
		mdl := mdl.(*InputField)
		mdlCopy = &InputField{
			LocationID: mdl.LocationID,
			TextTag:    mdl.TextTag,
			Value:      mdl.Value,
			pseudoID:   mdl.pseudoID,
		}
	case *Location:
		mdl := mdl.(*Location)
		mdlCopy = &Location{
			LocationID:     mdl.LocationID,
			BookNumber:     sql.NullInt32{Int32: mdl.BookNumber.Int32, Valid: mdl.BookNumber.Valid},
			ChapterNumber:  sql.NullInt32{Int32: mdl.ChapterNumber.Int32, Valid: mdl.ChapterNumber.Valid},
			DocumentID:     sql.NullInt32{Int32: mdl.DocumentID.Int32, Valid: mdl.DocumentID.Valid},
			Track:          sql.NullInt32{Int32: mdl.Track.Int32, Valid: mdl.Track.Valid},
			IssueTagNumber: mdl.IssueTagNumber,
			KeySymbol:      sql.NullString{String: mdl.KeySymbol.String, Valid: mdl.KeySymbol.Valid},
			MepsLanguage:   mdl.MepsLanguage,
			LocationType:   mdl.LocationType,
			Title:          sql.NullString{String: mdl.Title.String, Valid: mdl.Title.Valid},
			Specialty:      sql.NullString{String: mdl.Specialty.String, Valid: mdl.Specialty.Valid},
			Edition:        sql.NullString{String: mdl.Edition.String, Valid: mdl.Edition.Valid},
		}
	case *Note:
		mdl := mdl.(*Note)
		mdlCopy = &Note{
			NoteID:          mdl.NoteID,
			GUID:            mdl.GUID,
			UserMarkID:      sql.NullInt32{Int32: mdl.UserMarkID.Int32, Valid: mdl.UserMarkID.Valid},
			LocationID:      sql.NullInt32{Int32: mdl.LocationID.Int32, Valid: mdl.LocationID.Valid},
			Title:           sql.NullString{String: mdl.Title.String, Valid: mdl.Title.Valid},
			Content:         sql.NullString{String: mdl.Content.String, Valid: mdl.Content.Valid},
			LastModified:    mdl.LastModified,
			Created:         mdl.Created,
			BlockType:       mdl.BlockType,
			BlockIdentifier: sql.NullInt32{Int32: mdl.BlockIdentifier.Int32, Valid: mdl.BlockIdentifier.Valid},
		}
	case *Tag:
		mdl := mdl.(*Tag)
		mdlCopy = &Tag{
			TagID:   mdl.TagID,
			TagType: mdl.TagType,
			Name:    mdl.Name,
		}
	case *TagMap:
		mdl := mdl.(*TagMap)
		mdlCopy = &TagMap{
			TagMapID:       mdl.TagMapID,
			PlaylistItemID: sql.NullInt32{Int32: mdl.PlaylistItemID.Int32, Valid: mdl.PlaylistItemID.Valid},
			LocationID:     sql.NullInt32{Int32: mdl.LocationID.Int32, Valid: mdl.LocationID.Valid},
			NoteID:         sql.NullInt32{Int32: mdl.NoteID.Int32, Valid: mdl.NoteID.Valid},
			TagID:          mdl.TagID,
			Position:       mdl.Position,
		}
	case *UserMark:
		mdl := mdl.(*UserMark)
		mdlCopy = &UserMark{
			UserMarkID:   mdl.UserMarkID,
			ColorIndex:   mdl.ColorIndex,
			LocationID:   mdl.LocationID,
			StyleIndex:   mdl.StyleIndex,
			UserMarkGUID: mdl.UserMarkGUID,
			Version:      mdl.Version,
		}
	case *UserMarkBlockRange:
		mdl := mdl.(*UserMarkBlockRange)

		brSliceCopy := make([]*BlockRange, len(mdl.BlockRanges))
		for i, br := range mdl.BlockRanges {
			if br != nil {
				brSliceCopy[i] = MakeModelCopy(br).(*BlockRange)
			}
		}
		return &UserMarkBlockRange{
			UserMark:    MakeModelCopy(mdl.UserMark).(*UserMark),
			BlockRanges: brSliceCopy,
		}
	default:
		panic(fmt.Sprintf("Type %T is not supported for copying", mdl))
	}

	return mdlCopy
}

// prettyPrint prints the given fields of a Model as a table. If the field
// is empty, its omitted.
func prettyPrint(m Model, fields []string) string {
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)

Loop:
	for _, fieldName := range fields {
		field := reflect.ValueOf(m).Elem().FieldByName(fieldName)
		if !field.IsValid() {
			panic(fmt.Sprintf("Given struct does not contain field %s", fieldName))
		}
		switch field.Interface().(type) {
		case string:
			fmt.Fprintf(w, "\n%s:\t%s", fieldName, strings.ReplaceAll(wordwrap.WrapString(field.String(), 70), "\n", "\n\t"))
		case sql.NullString:
			if field.Field(1).Bool() == false {
				continue Loop
			}
			fmt.Fprintf(w, "\n%s:\t%s", fieldName, strings.ReplaceAll(wordwrap.WrapString(field.Field(0).String(), 70), "\n", "\n\t"))
		case int:
			fmt.Fprintf(w, "\n%s:\t%d", fieldName, field.Int())
		case sql.NullInt32:
			if field.Field(1).Bool() == false {
				continue Loop
			}
			fmt.Fprintf(w, "\n%s:\t%d", fieldName, field.Field(0).Int())
		default:
			panic(fmt.Sprintf("Unsupported type for field %s", fieldName))
		}
	}
	w.Flush()
	return buf.String()
}

// SortByUniqueKey sorts the given pointer to a slice of Model by UniqueKey,
// removes unnecessary nil-entries (except at position 0),
// and also updates the IDs accordingly. It tracks these changes
// by a map, for which the key represents the old ID,
// and value represents the new ID.
func SortByUniqueKey(slice interface{}) map[int]int {
	changes := map[int]int{}

	if reflect.TypeOf(slice).Kind() != reflect.Ptr || reflect.TypeOf(slice).Elem().Kind() != reflect.Slice {
		panic("Only pointer to slice is supported")
	}

	s := reflect.ValueOf(slice).Elem()

	// Sort by UniqueKey
	sort.Slice(s.Interface(), func(i, j int) bool {
		// Nil is always smaller than every other value
		jVal := s.Index(j)
		if jVal.IsNil() {
			return false
		}
		iVal := s.Index(i)
		if iVal.IsNil() {
			return true
		}

		iUQ := s.Index(i).MethodByName("UniqueKey").Call(nil)
		jUQ := s.Index(j).MethodByName("UniqueKey").Call(nil)
		return iUQ[0].String() < jUQ[0].String()
	})

	// If there are more than one nil values, remove all except one
	// (all nil values are located at the beginning)
	nilCount := 0
	for i := 0; i < s.Len(); i++ {
		if !s.Index(i).IsNil() {
			continue
		}
		nilCount++
	}
	if nilCount > 1 {
		s.Set(s.Slice(nilCount-1, s.Len()))
	}

	// Update IDs to their index
	for i := 0; i < s.Len(); i++ {
		elem := s.Index(i)
		if elem.IsNil() {
			continue
		}
		oldID := int(elem.MethodByName("ID").Call(nil)[0].Int())
		if oldID != i {
			changes[oldID] = i
			elem.MethodByName("SetID").Call([]reflect.Value{reflect.ValueOf(i)})
		}
	}

	return changes
}

// UpdateIDs updates a given ID (named by IDName) on the slice of *model.Model
// according to the given map, for which the key represents the old ID,
// and value represents the new ID.
func UpdateIDs(mdl interface{}, IDName string, changes map[int]int) {
	switch reflect.TypeOf(mdl).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(mdl)
		for i := 0; i < s.Len(); i++ {
			elem := s.Index(i)
			if elem.IsNil() {
				continue
			}

			field := elem.Elem().FieldByName(IDName)
			if !field.IsValid() {
				panic(fmt.Sprintf("Given struct does not contain field %s", IDName))
			}

			switch t := field.Interface().(type) {
			case int:
				if new, ok := changes[int(field.Int())]; ok {
					field.SetInt(int64(new))
				}
			case sql.NullInt32:
				val := field.Field(0)
				if new, ok := changes[int(val.Int())]; ok {
					val.SetInt(int64(new))
				}
			default:
				panic(fmt.Sprintf("Type %T of field %s is not supported!", t, IDName))
			}
		}
	default:
		panic("Only slices are supported!")
	}
}
