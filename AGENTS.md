# Agent context for go-library-merger

## What this project does

Merges two JW Library backup files (`.jwlibrary`) into one. A `.jwlibrary`
file is a zip archive containing a SQLite database (`userData.db`) and a
`manifest.json`.

## Key files

- `model/data/userData.db` — embedded SQLite template used when generating
  new backup output. Must match the current JW Library schema.
- `model/manifest.go` — defines `supportedSchemaVersionMin` and
  `supportedSchemaVersionMax`. Both must be updated together whenever JW
  Library releases a schema change.
- `model/Model.go` — contains `MakeModelCopy`, a hand-written deep copy
  switch. Every field of every model struct must be listed here explicitly.
  Easy to miss when adding fields to a struct.
- `model/Location.go` — the struct most likely to change when JW Library
  updates its schema.

## Responding to a JW Library schema update

When a new `.jwlibrary` backup appears with a higher schema version, the
typical workflow is:

1. Extract the backup (`unzip`) and compare `.schema` output from both
   SQLite files using `sqlite3`.
2. Update the affected model structs (`scanRow`, `UniqueKey`, `Equals`,
   `MarshalJSON`, `MakeModelCopy`).
3. Update `supportedSchemaVersionMin` and `supportedSchemaVersionMax` in
   `model/manifest.go`. The min should reflect the lowest schema version
   whose SQLite layout the code can actually read — if a column was added,
   old schemas without it are no longer readable and the min must move up.
4. Replace `model/data/userData.db` with the new empty database.
5. Migrate test fixtures (see below).

## Migrating test fixtures after a schema change

There are two kinds of test fixtures that need updating:

**Raw SQLite files** (`model/testdata/userData.db`,
`model/testdata/userData_withPlaylist.db`):
- Add columns directly: `sqlite3 <file> "ALTER TABLE ... ADD COLUMN ..."`

**Packaged backups** (`model/testdata/*.jwlibrary`):
- Extract with `unzip`, migrate the SQLite file, update `schemaVersion` in
  `manifest.json`, repack with `zip -j`.
- Also update the `schemaVersion` in `manifest_correct.json` if it is used
  by a validation test that calls `validateManifest()`.

After migrating, update any hardcoded hashes (e.g. in
`Test_createEmptySQLiteDB`) by running `sha256sum` on the new file.

## Manual end-to-end testing

```
go run . merge left.jwlibrary right.jwlibrary merged.jwlibrary \
  --bookmarks chooseLeft --markings chooseLeft \
  --notes chooseLeft --inputFields chooseLeft
```

Merging a backup with the empty `new.jwlibrary` is a useful round-trip
smoke test. Merging a backup with itself should produce identical row counts
with no duplicates.

## Running tests

```
go test ./...
```
