# Duplicate Note Detection and Removal Feature

## Overview
This fork adds duplicate note detection and removal functionality to go-jwlm, allowing users to identify and clean up duplicate notes in their JW Library backups.

## New Commands

### 1. `detect-duplicates`
Scans a backup file and displays all duplicate notes.

**Usage:**
```bash
go-jwlm detect-duplicates <backup-file>
```

**Example:**
```bash
go-jwlm detect-duplicates mybackup.jwlibrary
```

**Features:**
- Identifies duplicate notes based on matching title and content
- Groups duplicates together for easy review
- Displays detailed information for each duplicate (ID, GUID, title, content preview, dates)
- Shows total count of duplicate groups and notes

### 2. `remove-duplicates`
Interactively removes duplicate notes from a backup file.

**Usage:**
```bash
go-jwlm remove-duplicates <backup-file> <output-file>
```

**Example:**
```bash
go-jwlm remove-duplicates mybackup.jwlibrary cleaned.jwlibrary
```

**Features:**
- Detects duplicate notes automatically
- Presents each duplicate group with full note details
- Interactive selection - you choose which note to keep
- Option to "Keep all" (skip removing duplicates from a specific group)
- Removes all other duplicates from that group
- Properly reassigns Note IDs after removal
- Updates TagMap references to new Note IDs
- Saves cleaned backup to output file
- Preserves all non-duplicate notes

## Implementation Details

### Files Added
1. **`merger/DuplicateDetector.go`** - Core duplicate detection logic
   - `DetectDuplicateNotes()` - Scans notes and groups duplicates
   - `DuplicateGroup` - Struct representing a group of duplicate notes

2. **`cmd/detect_duplicates.go`** - CLI command for detecting duplicates
   - Imports backup file
   - Runs duplicate detection
   - Displays results in formatted tables

3. **`cmd/remove_duplicates.go`** - CLI command for removing duplicates
   - Imports backup file
   - Detects duplicates
   - Interactive selection using survey library (same as merge conflicts)
   - Filters out selected duplicates
   - Properly rebuilds Note array with correct ID indexing
   - Updates TagMap references to new Note IDs
   - Exports cleaned backup

### Duplicate Detection Algorithm
Notes are considered duplicates if they have:
- Identical title (or both empty)
- Identical content (or both empty)

The algorithm creates a signature from the title and content and groups notes with matching signatures.

### Design Decisions
1. **Reuses existing patterns** - The interactive selection interface follows the same pattern as the merge conflict resolution in the original code
2. **Non-destructive** - The remove-duplicates command requires an output file, preserving the original backup
3. **Conservative matching** - Only exact title+content matches are considered duplicates (not fuzzy matching)
4. **User control** - Interactive mode ensures users explicitly choose which notes to keep

## Building
```bash
go build -o go-jwlm
```

## Testing
To test the new functionality:

1. Create or find a backup file with duplicate notes
2. Run detection:
   ```bash
   ./go-jwlm detect-duplicates test.jwlibrary
   ```
3. Remove duplicates:
   ```bash
   ./go-jwlm remove-duplicates test.jwlibrary cleaned.jwlibrary
   ```

## Future Enhancements
Potential improvements:
- Fuzzy matching for similar (but not identical) notes
- Automatic mode (like merge conflict resolvers) - e.g., `--keepNewest`, `--keepOldest`
- Batch selection for duplicate groups
- Duplicate detection for other entities (bookmarks, markings, etc.)
- Preview mode showing what would be removed without actually removing it

## File Size Behavior

**Important:** The first time a backup is exported, the file size may appear slightly larger than the original due to SQLite database restructuring. This is normal and happens even without removing duplicates. The restructuring optimizes the database, and subsequent exports will maintain or reduce the size when duplicates are removed.

Example:
- Original backup: 3.29 MB
- After removing 128 duplicates: 3.57 MB (appears larger)
- However, without removing duplicates it would be ~3.7 MB
- Re-exporting the cleaned backup maintains the smaller size

The duplicate removal is working correctly - the file would be larger without it.

## Compatibility
- Uses existing dependencies (no new external packages required)
- Compatible with current go-jwlm architecture
- Follows existing code style and patterns
