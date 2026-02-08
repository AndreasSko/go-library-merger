package model

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"github.com/sergi/go-diff/diffmatchpatch"
	log "github.com/sirupsen/logrus"

	// Register SQLite driver
	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed data/userData.db
var userDataDatabaseFile []byte

//go:embed data/default_thumbnail.png
var defaultThumbnailFile []byte

const manifestFilename = "manifest.json"
const userDataFilename = "userData.db"
const defaultThumbnailFilename = "default_thumbnail.png"

// Database represents the JW Library database as a struct
type Database struct {
	BlockRange []*BlockRange
	Bookmark   []*Bookmark
	InputField []*InputField
	Location   []*Location
	Note       []*Note
	Tag        []*Tag
	TagMap     []*TagMap
	UserMark   []*UserMark

	// ContainsPlaylists indicates if the imported backup contains playlists.
	ContainsPlaylists bool
	// SkipPlaylists allows to skip prevention of merging if playlists exist in the database.
	// It is meant as a temporary workaround until merging of playlists is implemented.
	SkipPlaylists bool
	// TempDir is used for temporary files. If not set, os.TempDir() will be used.
	TempDir string
}

// FetchFromTable tries to fetch a entry with the given ID. If it can't find it
// or the entry is empty it returns nil.
func (db *Database) FetchFromTable(tableName string, id int) Model {
	if db == nil {
		return nil
	}

	table := reflect.ValueOf(db).Elem().FieldByName(tableName)
	if !table.IsValid() {
		panic(fmt.Sprintf("Table %s does not exist in Database", tableName))
	}

	if id >= table.Len() {
		return nil
	}
	if table.Index(id).IsNil() {
		return nil
	}

	return table.Index(id).Interface().(Model)
}

// PurgeTables removes all entries from the tables mentioned in the tables slice,
// which are named by the fields of the Database slice. If a table doesn't exist,
// an error will be returned.
func (db *Database) PurgeTables(tables []string) error {
	if db == nil {
		return fmt.Errorf("can't purge tables. Database is nil")
	}

	for _, tableName := range tables {
		if tableName == "" {
			continue
		}

		table := reflect.ValueOf(db).Elem().FieldByName(tableName)
		if !table.IsValid() {
			return fmt.Errorf("table %s does not exist in database", tableName)
		}
		table.Set(reflect.MakeSlice(table.Type(), 1, 1))
	}

	return nil
}

// MakeDatabaseCopy creates a deep copy of the given Database, so elements of
// the copy can be safely updated without affecting the original one.
func MakeDatabaseCopy(db *Database) *Database {
	newDB := &Database{}

	dbFields := reflect.ValueOf(db).Elem()
	for i := 0; i < dbFields.NumField(); i++ {
		field := dbFields.Field(i)
		if !field.CanInterface() {
			continue
		}

		cpField := reflect.ValueOf(newDB).Elem().Field(i)
		tp := field.Kind()
		switch tp {
		case reflect.Slice:
			cpSlice := reflect.MakeSlice(field.Type(), field.Len(), field.Len())

			for j := 0; j < field.Len(); j++ {
				elem := field.Index(j)

				if elem.IsNil() {
					continue
				}

				switch t := elem.Interface().(type) {
				case Model:
					cpSlice.Index(j).Set(reflect.ValueOf(MakeModelCopy(elem.Interface().(Model))))
				default:
					panic(fmt.Sprintf("Element type %T is not supported for copying", t))
				}
			}
			cpField.Set(cpSlice)
		case reflect.Bool:
			cpField.SetBool(field.Bool())
		case reflect.String:
			cpField.SetString(field.String())
		default:
			panic(fmt.Sprintf("Field type %T is not supported for copying", tp))
		}
	}

	return newDB
}

// Equals checks if all entries of a Database are equal.
func (db *Database) Equals(other *Database) bool {
	// Make copy of DBs so we can safely transform them if necessary
	dbCp := MakeDatabaseCopy(db)
	otherCp := MakeDatabaseCopy(other)

	// Sort all tables by UniqueKey and update IDs in other tables
	for _, db := range []*Database{dbCp, otherCp} {
		locIDChanges := SortByUniqueKey(&db.Location)
		UpdateIDs(db.Bookmark, "LocationID", locIDChanges)
		UpdateIDs(db.Bookmark, "PublicationLocationID", locIDChanges)
		UpdateIDs(db.InputField, "LocationID", locIDChanges)
		UpdateIDs(db.Note, "LocationID", locIDChanges)
		UpdateIDs(db.TagMap, "LocationID", locIDChanges)
		UpdateIDs(db.UserMark, "LocationID", locIDChanges)

		SortByUniqueKey(&db.Bookmark)
		SortByUniqueKey(&db.InputField)

		tagIDChanges := SortByUniqueKey(&db.Tag)
		UpdateIDs(db.TagMap, "TagID", tagIDChanges)

		umIDChanges := SortByUniqueKey(&db.UserMark)
		UpdateIDs(db.BlockRange, "UserMarkID", umIDChanges)
		UpdateIDs(db.Note, "UserMarkID", umIDChanges)

		SortByUniqueKey(&db.BlockRange)

		noteIDChanges := SortByUniqueKey(&db.Note)
		UpdateIDs(db.TagMap, "NoteID", noteIDChanges)

		SortByUniqueKey(&db.TagMap)
	}

	// Check if all entries are equal.
	dbFields := reflect.ValueOf(dbCp).Elem()
	otherFields := reflect.ValueOf(otherCp).Elem()
	for i := 0; i < dbFields.NumField(); i++ {
		dbField := dbFields.Field(i)
		otherField := otherFields.Field(i)

		tp := dbField.Kind()
		switch tp {
		case reflect.Slice:
			if !dbField.CanInterface() || !otherField.CanInterface() {
				continue
			}

			if dbField.Len() != otherField.Len() {
				fmt.Printf("Length of slices at index %d are not equal: %d vs %d\n", i, dbField.Len(), otherField.Len())
				return false
			}

			for j := 0; j < dbField.Len(); j++ {
				dElem := dbField.Index(j)
				oElem := otherField.Index(j)

				if dElem.IsNil() {
					if oElem.IsNil() {
						continue
					} else {
						return false
					}
				}

				if !dElem.MethodByName("Equals").Call([]reflect.Value{oElem})[0].Bool() {
					fmt.Println("Found different entries: ")
					left := spew.Sdump(dElem.Interface())
					right := spew.Sdump(oElem.Interface())
					fmt.Printf("%s \nvs\n %s", left, right)
					dmp := diffmatchpatch.New()
					diffs := dmp.DiffMain(left, right, true)
					fmt.Println("Diff:")
					fmt.Println(dmp.DiffPrettyText(diffs))
					return false
				}
			}
		case reflect.Bool:
			if dbFields.Field(i).Bool() != otherFields.Field(i).Bool() {
				return false
			}
		case reflect.String:
			if dbFields.Field(i).String() != otherFields.Field(i).String() {
				return false
			}
		default:
			panic(fmt.Sprintf("field type %T is not supported for checking equality", tp))
		}
	}

	return true
}

// ImportJWLBackup unzips a given JW Library Backup file and imports the
// included SQLite DB to the Database struct
func (db *Database) ImportJWLBackup(filename string) error {
	// Create tmp folder and place all files there
	tmp, err := os.MkdirTemp(db.TempDir, "go-jwlm")
	if err != nil {
		return errors.Wrap(err, "Error while creating temporary directory")
	}
	defer os.RemoveAll(tmp)

	r, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, file := range r.File {
		fileReader, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer fileReader.Close()

		path := filepath.Join(tmp, file.Name)
		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("failed to open target file: %w", err)
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return errors.Wrap(err, "Error while copying files from backup to temporary folder")
		}
	}

	// Import manifest
	path := filepath.Join(tmp, manifestFilename)
	manifest := manifest{}
	if err := manifest.importManifest(path); err != nil {
		return errors.Wrap(err, "Error while importing manifest")
	}

	// Make sure that we support this backup version
	if err := manifest.validateManifest(); err != nil {
		return fmt.Errorf("backup incompatible: %w", err)
	}

	// Fill the Database with actual data
	path = filepath.Join(tmp, manifest.UserDataBackup.DatabaseName)
	return db.importSQLite(path)
}

// importSQLite imports a given SQLite DB into the Database struct
func (db *Database) importSQLite(filename string) error {
	// Open SQLite file as immutable to avoid locks (and therefore speed up import)
	sqlite, err := sql.Open("sqlite3", filename+"?immutable=1")
	if err != nil {
		return errors.Wrap(err, "Error while opening SQLite database")
	}
	defer sqlite.Close()

	var wg sync.WaitGroup
	wg.Add(8)
	errors := make(chan error, 10)

	// Fill each table struct separately (did not find a DRYer solution yet..)
	go func() {
		mdl, err := fetchFromSQLite(sqlite, &BlockRange{})
		if err != nil {
			errors <- err
			wg.Done()
			return
		}
		db.BlockRange = BlockRange{}.MakeSlice(mdl)
		wg.Done()
	}()

	go func() {
		mdl, err := fetchFromSQLite(sqlite, &Bookmark{})
		if err != nil {
			errors <- err
			wg.Done()
			return
		}
		db.Bookmark = Bookmark{}.MakeSlice(mdl)
		wg.Done()
	}()

	go func() {
		mdl, err := fetchFromSQLite(sqlite, &InputField{})
		if err != nil {
			errors <- err
			wg.Done()
			return
		}
		db.InputField = InputField{}.MakeSlice(mdl)
		wg.Done()
	}()

	go func() {
		mdl, err := fetchFromSQLite(sqlite, &Location{})
		if err != nil {
			errors <- err
			wg.Done()
			return
		}
		db.Location = Location{}.MakeSlice(mdl)
		wg.Done()
	}()

	go func() {
		mdl, err := fetchFromSQLite(sqlite, &Note{})
		if err != nil {
			errors <- err
			wg.Done()
			return
		}
		db.Note = Note{}.MakeSlice(mdl)
		wg.Done()
	}()

	go func() {
		mdl, err := fetchFromSQLite(sqlite, &Tag{})
		if err != nil {
			errors <- err
			wg.Done()
			return
		}
		db.Tag = Tag{}.MakeSlice(mdl)
		wg.Done()
	}()

	go func() {
		mdl, err := fetchFromSQLite(sqlite, &TagMap{})
		if err != nil {
			errors <- err
			wg.Done()
			return
		}
		db.TagMap = TagMap{}.MakeSlice(mdl)
		wg.Done()
	}()

	go func() {
		mdl, err := fetchFromSQLite(sqlite, &UserMark{})
		if err != nil {
			errors <- err
			wg.Done()
			return
		}
		db.UserMark = UserMark{}.MakeSlice(mdl)
		wg.Done()
	}()

	wg.Wait()

	select {
	case err := <-errors:
		return err
	default:
	}

	// Make sure these tables are empty as we are not able to merge them yet.
	// Better to fail, than to risk losing data..
	emptyTables := []string{
		"IndependentMedia",
		"PlaylistItem",
		"PlaylistItemIndependentMediaMap",
		"PlaylistItemLocationMap",
		"PlaylistItemMarker",
		"PlaylistItemMarkerBibleVerseMap",
		"PlaylistItemMarkerParagraphMap",
	}
	for _, table := range emptyTables {
		count, err := getTableEntryCount(sqlite, table)
		if err != nil {
			return err
		}
		if count > 0 {
			db.ContainsPlaylists = true
		}
	}

	db.removePlaylists()

	if db.ContainsPlaylists && !db.SkipPlaylists {
		return fmt.Errorf("merging of playlists is not supported yet. Enable SkipPlaylists flag to skip this safety check")
	}

	return nil
}

// removePlaylists removes all playlists (represented as a Tag with type 2)
// and its items from the database. It indicates that the database contained
// playlists by setting the ContainedPlaylists field to true.
func (db *Database) removePlaylists() {
	for i, t := range db.Tag {
		if t != nil && t.TagType == 2 {
			db.ContainsPlaylists = true
			db.Tag[i] = nil
		}
	}

	for i, t := range db.TagMap {
		if t != nil && t.PlaylistItemID.Valid && t.PlaylistItemID.Int32 != 0 {
			db.ContainsPlaylists = true
			db.TagMap[i] = nil
		}
	}
}

// fetchFromSQLite fetches the entries for a given modelType and returns a slice
// of entries, for which the index corresponds to the ID in the SQLite DB
func fetchFromSQLite(sqlite *sql.DB, modelType Model) ([]Model, error) {
	// Create slice of correct size (number of entries)
	capacity, err := getSliceCapacity(sqlite, modelType)
	if err != nil {
		return nil, errors.Wrap(err, "Could not determine number of entries in SQLite database")
	}
	result := make([]Model, capacity)

	rows, err := sqlite.Query(fmt.Sprintf("SELECT * FROM %s", modelType.tableName()))
	if err != nil {
		return nil, errors.Wrap(err, "Error while querying SQLite database")
	}

	// Put entries in slice with the index coresponding to the ID in the SQLite DB
	i := 1
	defer rows.Close()
	for rows.Next() {
		var m Model
		switch tp := modelType.(type) {
		case *BlockRange:
			m = &BlockRange{}
		case *Bookmark:
			m = &Bookmark{}
		case *InputField:
			m = &InputField{pseudoID: i}
		case *Location:
			m = &Location{}
		case *Note:
			m = &Note{}
		case *Tag:
			m = &Tag{}
		case *TagMap:
			m = &TagMap{}
		case *UserMark:
			m = &UserMark{}
		default:
			panic(fmt.Sprintf("Fetching %T is not supported!", tp))
		}
		mn, err := m.scanRow(rows)
		if err != nil {
			// For some reason a row might contain NULL entries, even though the schema
			// shouldn't allow this. Instead of failing the whole import, we can simply skip
			// this entry as the data anyway wouldn't be valid.
			if !isNullableMismatch(err, rows) {
				return nil, errors.Wrapf(err, "Error while scanning row for %T", modelType)
			}
			fmt.Println(err)
			log.Warnf("Nullable mismatch in %T at index %d detected. Skipping entry", m, i)
			i++
			continue
		}
		result[mn.ID()] = mn
		i++
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "Error while scanning results for %T from SQLite database", modelType)
	}

	return result, nil
}

// isNullableMismatch checks if a given error is due to a NULL entry in a column that only allows
// non-NULL entries. If this is the case and no other schema mismatch is detected, true is returned.
func isNullableMismatch(err error, rows *sql.Rows) bool {
	if err == nil || rows == nil {
		return false
	}

	re := regexp.MustCompile(`Scan error on column index (\d+), name "(\w+)": converting NULL to (\w+) is unsupported`)
	matches := re.FindStringSubmatch(err.Error())
	if len(matches) != 4 {
		return false
	}

	index, err := strconv.ParseInt(matches[1], 0, 64)
	if err != nil {
		return false
	}

	ct, err := rows.ColumnTypes()
	if err != nil {
		return false
	}

	if len(ct) <= int(index) {
		return false
	}

	column := ct[index]
	if column == nil {
		return false
	}

	if column.Name() != matches[2] {
		return false
	}

	if typeName, ok := dbTypeToGoType[column.DatabaseTypeName()]; !ok || typeName != matches[3] {
		return false
	}

	if n, _ := column.Nullable(); n {
		return false
	}

	return true
}

// getTableEntryCount returns the number of entries in a given table
func getTableEntryCount(sqlite *sql.DB, tableName string) (int, error) {
	var count int
	err := sqlite.QueryRow(fmt.Sprintf("SELECT Count(*) FROM %s", tableName)).Scan(&count)
	if err != nil {
		return 0, errors.Wrapf(err, "Error while determing entry count of table %s", tableName)
	}

	return count, nil
}

// getSliceCapacity determines the needed capacity for a slice from a table
// by looking at the highest ID in the DB. If the table does not have a ID
// column, it will simply count the number of entries.
func getSliceCapacity(sqlite *sql.DB, modelType Model) (int, error) {
	var query string
	if modelType.idName() != "" {
		query = fmt.Sprintf("SELECT %s FROM %s ORDER BY %s DESC LIMIT 1",
			modelType.idName(), modelType.tableName(), modelType.idName())
	} else {
		query = fmt.Sprintf("SELECT COUNT(*) FROM %s", modelType.tableName())
	}
	row, err := sqlite.Query(query)
	if err != nil {
		return 0, err
	}
	defer row.Close()

	capacity := 0
	for row.Next() {
		if err := row.Scan(&capacity); err != nil {
			return 0, err
		}
	}

	// Index in DB starts with 1, so 0 is always nil
	return capacity + 1, nil
}

// ExportJWLBackup creates a .jwlibrary backup file out of a Database{} struct
func (db *Database) ExportJWLBackup(filename string) error {
	// Create tmp folder and place all files there
	tmp, err := os.MkdirTemp(db.TempDir, "go-jwlm")
	if err != nil {
		return errors.Wrap(err, "Error while creating temporary directory")
	}
	defer os.RemoveAll(tmp)

	// Create userData.db
	dbPath := filepath.Join(tmp, userDataFilename)
	if err := db.saveToNewSQLite(dbPath); err != nil {
		return errors.Wrap(err, "Could not create SQLite database for exporting")
	}

	// VACUUM to defragment and optimize the database file
	if err := vacuumDatabase(dbPath); err != nil {
		return errors.Wrap(err, "Could not vacuum SQLite database")
	}

	// Create manifest.json
	manifestPath := filepath.Join(tmp, manifestFilename)
	mfst, err := generateManifest("go-jwlm", dbPath)
	if err != nil {
		return errors.Wrap(err, "Error while generating manifest")
	}
	if err := mfst.exportManifest(manifestPath); err != nil {
		return errors.Wrap(err, "Error while creating manifest.json")
	}

	defaultThumbnailPath := filepath.Join(tmp, defaultThumbnailFilename)
	if err := os.WriteFile(defaultThumbnailPath, defaultThumbnailFile, 0644); err != nil {
		return fmt.Errorf("writing default thumbnail to %s: %w", defaultThumbnailPath, err)
	}

	// Store files in .jwlibrary (zip)-file
	files := []string{dbPath, manifestPath, defaultThumbnailPath}
	if err := zipFiles(filename, files); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error while storing files in zip archive %s", filename))
	}

	return nil
}

// saveToNewSQLite creates a new SQLite database with the JW Library scheme
// and saves all entries of the Database{} struct to it
func (db *Database) saveToNewSQLite(filename string) error {
	if err := createEmptySQLiteDB(filename); err != nil {
		return errors.Wrap(err, "Error while creating new empty SQLite database")
	}

	sqlite, err := sql.Open("sqlite3", filename)
	if err != nil {
		return errors.Wrap(err, "Error while opening SQLite database")
	}
	defer sqlite.Close()

	// Set journal mode to DELETE for compatibility (matches JW Library backups)
	// Note: page_size is already 4096 in the embedded database template
	// Note: auto_vacuum is intentionally left as NONE (0) to match JW Library backups
	//       and avoid overhead pages that reduce compressibility
	if _, err := sqlite.Exec("PRAGMA journal_mode = DELETE"); err != nil {
		log.Debugf("Warning: Failed to set journal_mode: %v", err)
	}

	// For every field of the Database{} struct, create a []model slice
	// and use it to insert its entries to the new SQLite DB
	dbFields := reflect.ValueOf(db).Elem()
	for j := 0; j < dbFields.NumField(); j++ {
		if dbFields.Field(j).Kind() != reflect.Slice {
			continue
		}
		slice := dbFields.Field(j).Interface()
		mdl, err := MakeModelSlice(slice)
		if err != nil {
			return err
		}
		if err := insertEntries(sqlite, mdl); err != nil {
			return errors.Wrapf(err, "Error while inserting entries of field %d", j)
		}
	}

	return nil
}

// vacuumDatabase performs a VACUUM on the SQLite database file to defragment
// and optimize it. This is called separately after saveToNewSQLite to allow
// testing with and without VACUUM.
func vacuumDatabase(filename string) error {
	sqlite, err := sql.Open("sqlite3", filename)
	if err != nil {
		return errors.Wrap(err, "Error while opening SQLite database for VACUUM")
	}
	defer sqlite.Close()

	_, err = sqlite.Exec("VACUUM")
	if err != nil {
		return errors.Wrap(err, "Error while vacuuming SQLite DB")
	}

	return nil
}

// insertEntries INSERTs entries of []model into a given SQLite database.
// It does it by dynamically parsing all fields of a struct implementing
// model using reflection and creating a query for SQLite out of it.
func insertEntries(sqlite *sql.DB, m []Model) error {
	if len(m) == 0 {
		return nil
	}

	// Figure out tableName and rowCount. As we don't know for sure,
	// which entry will be nil-pointer, we just try until we find a non-empty
	// one and call the functions there.
	tableName := ""
	rowCount := 0
	foundEntry := false
	for _, mdl := range m {
		if reflect.ValueOf(mdl).Elem().IsValid() {
			tableName = mdl.tableName()
			foundEntry = true

			// Count number of fields that don't have the "ignore" tag set
			reflTypes := reflect.TypeOf(mdl).Elem()
			for j := 0; j < reflTypes.NumField(); j++ {
				if _, ignore := reflTypes.Field(j).Tag.Lookup("ignore"); ignore {
					continue
				}
				rowCount++
			}
			break
		}
	}
	// If slice is empty, we don't need to continue
	if !foundEntry {
		return nil
	}

	tx, err := sqlite.Begin()
	if err != nil {
		return err
	}

	// Dynamically add all column-names of the struct to the query
	query := fmt.Sprintf("INSERT INTO %s VALUES (", tableName)
	for i := 0; i < rowCount; i++ {
		query += "?"
		if i+1 < rowCount {
			query += ", "
		}
	}
	query += ")"

	stmt, err := tx.Prepare(query)
	if err != nil {
		return errors.Wrapf(err, "Error while preparing query %s", query)
	}
	defer stmt.Close()

	for _, entry := range m {
		// Prepare struct for ingestion with stmt.Exec
		values := make([]interface{}, rowCount)
		reflValues := reflect.ValueOf(entry).Elem()
		reflTypes := reflect.TypeOf(entry).Elem()

		// Check if entry is actually a nil-pointer and shouldn't be considered
		if !reflValues.IsValid() {
			continue
		}

		// Add all fields of the struct to the values slice, so we can ingest them later
		for j := 0; j < reflValues.NumField(); j++ {
			// If struct field has `ignore` tag then skip it
			if _, ignore := reflTypes.Field(j).Tag.Lookup("ignore"); ignore {
				continue
			}
			v := reflValues.Field(j).Interface()
			values[j] = v
		}

		if _, err := stmt.Exec(values...); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Could not insert entry %v", entry))
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "Error while commiting entries")
	}

	return nil
}

// createEmptySQLiteDB creates a new SQLite database at filename with the base userData.db from JWLibrary
func createEmptySQLiteDB(filename string) error {
	if err := os.WriteFile(filename, userDataDatabaseFile, 0644); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Error while saving new SQLite database at %s", filename))
	}

	return nil
}

// maps a DatabaseTypeName to a go type name
var dbTypeToGoType = map[string]string{
	"INTEGER": "int",
	"TEXT":    "string",
}
