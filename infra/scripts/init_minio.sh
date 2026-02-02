#!/bin/sh
set -e

echo "[minio-init] endpoint=http://${MINIO_ENDPOINT} bucket=${MINIO_BUCKET} alias=local"

mc alias set local "http://${MINIO_ENDPOINT}" "${MINIO_ACCESS_KEY}" "${MINIO_SECRET_KEY}"

if mc ls "local/${MINIO_BUCKET}" >/dev/null 2>&1; then
  echo "[minio-init] bucket ${MINIO_BUCKET} already exists"
else
  echo "[minio-init] creating bucket ${MINIO_BUCKET}"
  mc mb "local/${MINIO_BUCKET}"
fi

echo "[minio-init] setting bucket policy to private"
mc anonymous set none "local/${MINIO_BUCKET}"

echo "[minio-init] Bucket ensured + policy internal (anonymous=none)"
