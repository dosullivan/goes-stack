#!/bin/bash

# GOES Data Upload Script for MinIO
# This script syncs all processed satellite data to MinIO and cleans up local files
# Preserves directory structure: goes19/, goes18/, himawari8/, emwin/, nws/

set -euo pipefail

# Configuration from environment variables
MINIO_ENDPOINT=${MINIO_ENDPOINT:-localhost:9000}
MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY:-minioadmin}
MINIO_SECRET_KEY=${MINIO_SECRET_KEY:-minioadmin}
BUCKET_NAME=${BUCKET_NAME:-goes-data}
SOURCE_DIR="/data"
LOG_PREFIX="[$(date '+%Y-%m-%d %H:%M:%S')]"
ALLOW_REMOTE_DELETIONS=${ALLOW_REMOTE_DELETIONS:-false}

echo "$LOG_PREFIX Starting upload process..."

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

# Count files before upload (including all data types)
file_count=$(find "$SOURCE_DIR" -type f \( -name "*.jpg" -o -name "*.png" -o -name "*.gif" -o -name "*.txt" -o -name "*.TXT" -o -name "*.nc" \) | wc -l)
echo "$LOG_PREFIX Found $file_count data files to process"

if [ "$file_count" -eq 0 ]; then
    echo "$LOG_PREFIX No files to upload"
    exit 0
fi

# Sync files to MinIO (safe by default: do NOT delete remote files)
echo "$LOG_PREFIX Syncing files to MinIO bucket: $BUCKET_NAME"
if [ "$ALLOW_REMOTE_DELETIONS" = "true" ]; then
    echo "$LOG_PREFIX WARNING: Remote deletions are ENABLED (ALLOW_REMOTE_DELETIONS=true)"
    mc mirror --remove --exclude "*.tmp" --exclude "*.partial" "$SOURCE_DIR/" "minio/$BUCKET_NAME/"
else
    mc mirror --exclude "*.tmp" --exclude "*.partial" "$SOURCE_DIR/" "minio/$BUCKET_NAME/"
fi

# Verify upload success and cleanup
echo "$LOG_PREFIX Cleaning up successfully uploaded files..."
cleanup_count=0

# Get list of files in bucket
mc ls "minio/$BUCKET_NAME/" --recursive | while read -r line; do
    # Extract filename from mc ls output (last column)
    remote_file=$(echo "$line" | awk '{print $NF}')
    local_file="$SOURCE_DIR/$remote_file"
    
    # Remove local file if it exists and was successfully uploaded
    if [[ -f "$local_file" ]]; then
        echo "$LOG_PREFIX Removing uploaded file: $local_file"
        rm -f "$local_file"
        cleanup_count=$((cleanup_count + 1))
    fi
done

# Remove empty directories
echo "$LOG_PREFIX Cleaning up empty directories..."
find "$SOURCE_DIR" -type d -empty -delete 2>/dev/null || true

# Get final counts
remaining_files=$(find "$SOURCE_DIR" -type f \( -name "*.jpg" -o -name "*.png" -o -name "*.gif" -o -name "*.txt" -o -name "*.TXT" -o -name "*.nc" \) | wc -l)
bucket_files=$(mc ls "minio/$BUCKET_NAME/" --recursive | wc -l)

echo "$LOG_PREFIX Upload completed successfully"
echo "$LOG_PREFIX Files remaining locally: $remaining_files"
echo "$LOG_PREFIX Total files in bucket: $bucket_files"