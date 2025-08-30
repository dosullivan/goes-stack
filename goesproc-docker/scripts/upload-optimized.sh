#!/bin/bash

# Optimized GOES Data Upload Script for MinIO
# Uses batch operations and parallel processing for faster uploads

set -euo pipefail

# Configuration from environment variables
MINIO_ENDPOINT=${MINIO_ENDPOINT:-localhost:9000}
MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY:-minioadmin}
MINIO_SECRET_KEY=${MINIO_SECRET_KEY:-minioadmin}
BUCKET_NAME=${BUCKET_NAME:-goes-data}
SOURCE_DIR="/data"
LOG_PREFIX="[$(date '+%Y-%m-%d %H:%M:%S')]"
ALLOW_REMOTE_DELETIONS=${ALLOW_REMOTE_DELETIONS:-false}

echo "$LOG_PREFIX Starting optimized upload process..."

# Run EMWIN preprocessing if the script exists
PREPROCESS_SCRIPT="/scripts/preprocess_emwin_fixed.sh"
if [[ -f "$PREPROCESS_SCRIPT" ]] && [[ -x "$PREPROCESS_SCRIPT" ]]; then
    echo "$LOG_PREFIX Running EMWIN file preprocessing..."
    "$PREPROCESS_SCRIPT" || echo "$LOG_PREFIX WARNING: EMWIN preprocessing failed"
fi

# Verify MinIO connection
if ! mc ls minio/ > /dev/null 2>&1; then
    echo "$LOG_PREFIX ERROR: Cannot connect to MinIO at $MINIO_ENDPOINT"
    exit 1
fi

# Create bucket if it doesn't exist
if ! mc ls "minio/$BUCKET_NAME" > /dev/null 2>&1; then
    echo "$LOG_PREFIX Creating bucket: $BUCKET_NAME"
    mc mb "minio/$BUCKET_NAME"
    mc anonymous set public "minio/$BUCKET_NAME"
fi

# Count files before upload
file_count=$(find "$SOURCE_DIR" -type f \( -name "*.jpg" -o -name "*.png" -o -name "*.gif" -o -name "*.txt" -o -name "*.TXT" -o -name "*.nc" \) 2>/dev/null | wc -l)
echo "$LOG_PREFIX Found $file_count data files to process"

if [ "$file_count" -eq 0 ]; then
    echo "$LOG_PREFIX No files to upload"
    exit 0
fi

# Use mirror with optimizations
echo "$LOG_PREFIX Syncing files to MinIO bucket: $BUCKET_NAME"

# Mirror options:
# --overwrite: Force overwrite of existing objects
# --preserve: Preserve file attributes
# --exclude: Skip temporary files
if [ "$ALLOW_REMOTE_DELETIONS" = "true" ]; then
    echo "$LOG_PREFIX WARNING: Remote deletions are ENABLED"
    mc mirror --remove --overwrite --preserve \
        --exclude "*.tmp" --exclude "*.partial" \
        "$SOURCE_DIR/" "minio/$BUCKET_NAME/"
else
    mc mirror --overwrite --preserve \
        --exclude "*.tmp" --exclude "*.partial" \
        "$SOURCE_DIR/" "minio/$BUCKET_NAME/"
fi

# Faster cleanup using find with -delete
echo "$LOG_PREFIX Cleaning up successfully uploaded files..."

# Get list of successfully uploaded files and remove them
# This is faster than checking each file individually
find "$SOURCE_DIR" -type f \( -name "*.jpg" -o -name "*.png" -o -name "*.gif" -o -name "*.txt" -o -name "*.TXT" -o -name "*.nc" \) -print0 | \
while IFS= read -r -d '' local_file; do
    # Extract relative path
    rel_path="${local_file#$SOURCE_DIR/}"
    
    # Check if file exists in MinIO (faster stat check)
    if mc stat "minio/$BUCKET_NAME/$rel_path" >/dev/null 2>&1; then
        rm -f "$local_file"
    fi
done

# Remove empty directories
find "$SOURCE_DIR" -type d -empty -delete 2>/dev/null || true

# Get final counts
remaining_files=$(find "$SOURCE_DIR" -type f \( -name "*.jpg" -o -name "*.png" -o -name "*.gif" -o -name "*.txt" -o -name "*.TXT" -o -name "*.nc" \) 2>/dev/null | wc -l)

echo "$LOG_PREFIX Upload completed successfully"
echo "$LOG_PREFIX Files remaining locally: $remaining_files"
echo "$LOG_PREFIX Files uploaded this cycle: $((file_count - remaining_files))"