#!/bin/bash

# Define variables
SOURCE_DIR="goes19/"
BUCKET="s3://goes-19/"
LOG_FILE="/home/daniel/upload-logs/s3.log"

# Sync files to S3
/usr/local/bin/s3cmd sync --no-check-md5 --recursive --acl-public --no-delete-removed "$SOURCE_DIR" "$BUCKET" --exclude 'm1/*' -v >> "$LOG_FILE"

# Remove successfully uploaded files
/usr/local/bin/s3cmd ls "$BUCKET" --recursive | while read -r LINE; do
  # Extract the file path from the bucket listing
  REMOTE_FILE=$(echo "$LINE" | awk '{print $4}')

  # Remove the bucket name prefix to get the relative path
  RELATIVE_PATH="${REMOTE_FILE#s3://goes-19/}"

  # Construct the full local file path
  LOCAL_FILE="${SOURCE_DIR}${RELATIVE_PATH}"

  # Check if the file exists locally, and remove it if it does
  if [[ -f "$LOCAL_FILE" ]]; then
    echo "Removing: $LOCAL_FILE"
    rm -f "$LOCAL_FILE"
  fi
done

# Remove empty directories
find "$SOURCE_DIR" -type d -empty -delete