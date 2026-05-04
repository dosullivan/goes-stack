#!/bin/bash

# GOES data upload script — talks to any S3-compatible backend (RustFS, MinIO, AWS S3)
# via the aws-cli.

set -euo pipefail

S3_ENDPOINT=${S3_ENDPOINT:-localhost:9000}
S3_ACCESS_KEY=${S3_ACCESS_KEY:-${MINIO_ACCESS_KEY:-minioadmin}}
S3_SECRET_KEY=${S3_SECRET_KEY:-${MINIO_SECRET_KEY:-minioadmin}}
S3_USE_SSL=${S3_USE_SSL:-false}
S3_REGION=${S3_REGION:-us-east-1}
BUCKET_NAME=${BUCKET_NAME:-goes-data}
SOURCE_DIR="/data"
LOG_PREFIX="[$(date '+%Y-%m-%d %H:%M:%S')]"
ALLOW_REMOTE_DELETIONS=${ALLOW_REMOTE_DELETIONS:-false}

if [ "$S3_USE_SSL" = "true" ]; then
    ENDPOINT_URL="https://$S3_ENDPOINT"
else
    ENDPOINT_URL="http://$S3_ENDPOINT"
fi

export AWS_ACCESS_KEY_ID="$S3_ACCESS_KEY"
export AWS_SECRET_ACCESS_KEY="$S3_SECRET_KEY"
export AWS_DEFAULT_REGION="$S3_REGION"

aws_s3() {
    aws --endpoint-url "$ENDPOINT_URL" s3 "$@"
}
aws_s3api() {
    aws --endpoint-url "$ENDPOINT_URL" s3api "$@"
}

echo "$LOG_PREFIX Starting upload process..."

# Run EMWIN preprocessing if the script exists
PREPROCESS_SCRIPT="/scripts/preprocess_emwin.sh"
if [[ -f "$PREPROCESS_SCRIPT" ]] && [[ -x "$PREPROCESS_SCRIPT" ]]; then
    echo "$LOG_PREFIX Running EMWIN file preprocessing..."
    "$PREPROCESS_SCRIPT" || echo "$LOG_PREFIX WARNING: EMWIN preprocessing failed"
fi

# Verify connection
if ! aws_s3 ls > /dev/null 2>&1; then
    echo "$LOG_PREFIX ERROR: Cannot connect to S3 endpoint at $ENDPOINT_URL"
    exit 1
fi

# Create bucket and set public-read policy if it doesn't exist
if ! aws_s3api head-bucket --bucket "$BUCKET_NAME" >/dev/null 2>&1; then
    echo "$LOG_PREFIX Creating bucket: $BUCKET_NAME"
    aws_s3 mb "s3://$BUCKET_NAME"
    aws_s3api put-bucket-policy --bucket "$BUCKET_NAME" --policy "{
        \"Version\": \"2012-10-17\",
        \"Statement\": [{
            \"Effect\": \"Allow\",
            \"Principal\": \"*\",
            \"Action\": [\"s3:GetObject\"],
            \"Resource\": [\"arn:aws:s3:::$BUCKET_NAME/*\"]
        }]
    }"
fi

# Count files before upload
file_count=$(find "$SOURCE_DIR" -type f \( -name "*.jpg" -o -name "*.png" -o -name "*.gif" -o -name "*.txt" -o -name "*.TXT" -o -name "*.nc" \) 2>/dev/null | wc -l)
echo "$LOG_PREFIX Found $file_count data files to process"

if [ "$file_count" -eq 0 ]; then
    echo "$LOG_PREFIX No files to upload"
    exit 0
fi

echo "$LOG_PREFIX Syncing files to bucket: $BUCKET_NAME"

SYNC_ARGS=(
    --exclude "*"
    --include "*.jpg" --include "*.png" --include "*.gif"
    --include "*.txt" --include "*.TXT" --include "*.nc"
)
if [ "$ALLOW_REMOTE_DELETIONS" = "true" ]; then
    echo "$LOG_PREFIX WARNING: Remote deletions are ENABLED"
    SYNC_ARGS+=(--delete)
fi

aws_s3 sync "$SOURCE_DIR/" "s3://$BUCKET_NAME/" "${SYNC_ARGS[@]}"

# Cleanup: delete local files that have a matching object remotely
echo "$LOG_PREFIX Cleaning up successfully uploaded files..."

find "$SOURCE_DIR" -type f \( -name "*.jpg" -o -name "*.png" -o -name "*.gif" -o -name "*.txt" -o -name "*.TXT" -o -name "*.nc" \) -print0 | \
while IFS= read -r -d '' local_file; do
    rel_path="${local_file#$SOURCE_DIR/}"
    if aws_s3api head-object --bucket "$BUCKET_NAME" --key "$rel_path" >/dev/null 2>&1; then
        rm -f "$local_file"
    fi
done

# Remove empty directories
find "$SOURCE_DIR" -type d -empty -delete 2>/dev/null || true

remaining_files=$(find "$SOURCE_DIR" -type f \( -name "*.jpg" -o -name "*.png" -o -name "*.gif" -o -name "*.txt" -o -name "*.TXT" -o -name "*.nc" \) 2>/dev/null | wc -l)

echo "$LOG_PREFIX Upload completed successfully"
echo "$LOG_PREFIX Files remaining locally: $remaining_files"
echo "$LOG_PREFIX Files uploaded this cycle: $((file_count - remaining_files))"
