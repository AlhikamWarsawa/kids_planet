#!/usr/bin/env sh
set -eu

ALIAS_NAME="${MINIO_ALIAS_NAME:-local}"

ENDPOINT="$MINIO_ENDPOINT"
case "$ENDPOINT" in
  http://*|https://*) ;;
  *) ENDPOINT="http://$ENDPOINT" ;;
esac

echo "[minio-init] endpoint=$ENDPOINT bucket=$MINIO_BUCKET alias=$ALIAS_NAME"

i=0
until mc alias set "$ALIAS_NAME" "$ENDPOINT" "$MINIO_ACCESS_KEY" "$MINIO_SECRET_KEY" >/dev/null 2>&1; do
  i=$((i+1))
  if [ "$i" -ge 60 ]; then
    echo "[minio-init] ERROR: MinIO not reachable after 60 tries"
    exit 1
  fi
  echo "[minio-init] waiting for MinIO... ($i/60)"
  sleep 2
done

echo "[minio-init] MinIO reachable"

mc mb --ignore-existing "$ALIAS_NAME/$MINIO_BUCKET" >/dev/null

mc anonymous set none "$ALIAS_NAME/$MINIO_BUCKET" >/dev/null

echo "[minio-init] Bucket ensured + policy internal (anonymous=none)"
