#!/bin/sh
set -e

echo "[minio-init] starting..."

: "${MINIO_ENDPOINT:?MINIO_ENDPOINT is required}"
: "${MINIO_ACCESS_KEY:?MINIO_ACCESS_KEY is required}"
: "${MINIO_SECRET_KEY:?MINIO_SECRET_KEY is required}"
: "${MINIO_BUCKET:?MINIO_BUCKET is required}"

echo "[minio-init] waiting for MinIO at http://$MINIO_ENDPOINT ..."
i=0
until mc alias set local "http://$MINIO_ENDPOINT" "$MINIO_ACCESS_KEY" "$MINIO_SECRET_KEY" >/dev/null 2>&1; do
  i=$((i+1))
  if [ "$i" -ge 60 ]; then
    echo "[minio-init] ERROR: timeout waiting for MinIO"
    exit 1
  fi
  sleep 1
done

echo "[minio-init] alias set successfully"

echo "[minio-init] ensure bucket exists: $MINIO_BUCKET"
mc mb --ignore-existing "local/$MINIO_BUCKET"

echo "[minio-init] set bucket to public download (anonymous read)"
mc anonymous set download "local/$MINIO_BUCKET" || true

echo "[minio-init] verify policy"
mc anonymous get "local/$MINIO_BUCKET" || true

echo "[minio-init] done"
