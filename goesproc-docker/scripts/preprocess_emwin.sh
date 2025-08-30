#!/bin/bash

# EMWIN File Preprocessing Script
# Reorganizes EMWIN files from 1969-12-31 folder to correct date folders BEFORE upload
# This runs locally on the filesystem before syncing to MinIO

set -euo pipefail

LOG_PREFIX="[$(date '+%Y-%m-%d %H:%M:%S')] [EMWIN-PREPROCESS]"
SOURCE_DIR="/data"
EMWIN_PATH="emwin"
MISPLACED_FOLDER="1969-12-31"
MISPLACED_DIR="$SOURCE_DIR/$EMWIN_PATH/$MISPLACED_FOLDER"

# Check if the misplaced folder exists
if [[ ! -d "$MISPLACED_DIR" ]]; then
    echo "$LOG_PREFIX No $MISPLACED_FOLDER folder found, nothing to reorganize"
    exit 0
fi

echo "$LOG_PREFIX Starting EMWIN file reorganization (local)..."

# Count files to process
TOTAL_FILES=$(find "$MISPLACED_DIR" -name "*.TXT" -o -name "*.txt" 2>/dev/null | wc -l)

if [ "$TOTAL_FILES" -eq 0 ]; then
    echo "$LOG_PREFIX No files found in $MISPLACED_FOLDER folder"
    # Remove empty directory
    rmdir "$MISPLACED_DIR" 2>/dev/null || true
    exit 0
fi

echo "$LOG_PREFIX Found $TOTAL_FILES files to reorganize"

MOVED=0
FAILED=0

# Process all .TXT files in the misplaced folder
for FILE_PATH in "$MISPLACED_DIR"/*.TXT "$MISPLACED_DIR"/*.txt; do
    # Skip if no files match
    [[ -e "$FILE_PATH" ]] || continue
    
    FILENAME=$(basename "$FILE_PATH")
    
    # Extract timestamp from filename
    # Format: A_ABCN01KWBC300115_C_KWIN_20250830011502_273154-2-STPTPTCN.TXT
    # The timestamp is: 20250830011502 (YYYYMMDDHHMMSS)
    if [[ "$FILENAME" =~ _([0-9]{14})_ ]]; then
        TIMESTAMP="${BASH_REMATCH[1]}"
        
        # Extract date from timestamp (YYYYMMDD)
        DATE_PART="${TIMESTAMP:0:8}"
        
        # Convert to YYYY-MM-DD format
        YEAR="${DATE_PART:0:4}"
        MONTH="${DATE_PART:4:2}"
        DAY="${DATE_PART:6:2}"
        TARGET_DATE="$YEAR-$MONTH-$DAY"
        
        # Validate date (basic check for reasonable years)
        if [[ "$YEAR" -ge 2024 ]] && [[ "$YEAR" -le 2030 ]]; then
            # Create target directory if it doesn't exist
            TARGET_DIR="$SOURCE_DIR/$EMWIN_PATH/$TARGET_DATE"
            mkdir -p "$TARGET_DIR"
            
            # Move the file
            TARGET_PATH="$TARGET_DIR/$FILENAME"
            if mv "$FILE_PATH" "$TARGET_PATH" 2>/dev/null; then
                ((MOVED++))
                if (( MOVED % 100 == 0 )); then
                    echo "$LOG_PREFIX Progress: $MOVED/$TOTAL_FILES files moved..."
                fi
            else
                echo "$LOG_PREFIX Failed to move: $FILENAME"
                ((FAILED++))
            fi
        else
            echo "$LOG_PREFIX Invalid date in filename: $FILENAME (extracted: $TARGET_DATE)"
            ((FAILED++))
        fi
    else
        # Can't parse timestamp, leave file in place for manual review
        echo "$LOG_PREFIX Cannot parse timestamp from: $FILENAME"
        ((FAILED++))
    fi
done

echo "$LOG_PREFIX Reorganization complete"
echo "$LOG_PREFIX Successfully moved: $MOVED files"
echo "$LOG_PREFIX Failed/skipped: $FAILED files"

# Remove the 1969-12-31 directory if it's empty
if [[ -d "$MISPLACED_DIR" ]]; then
    rmdir "$MISPLACED_DIR" 2>/dev/null && echo "$LOG_PREFIX Removed empty $MISPLACED_FOLDER folder" || true
fi

echo "$LOG_PREFIX Done!"