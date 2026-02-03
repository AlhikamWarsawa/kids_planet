# Runbook

## Game assets via MinIO + Nginx

Public URL format:

```
/games/{id}/current/index.html
```

### Troubleshooting

403 from MinIO (AccessDenied)
- Bucket policy is private. Verify anonymous read on the bucket:
  - `mc anonymous get local/games`
  - `mc anonymous set download local/games`
- Confirm the bucket name is `games` and the URL is `/games/...` (path-style access).

404 asset not found
- Object is missing: `mc ls local/games/{id}/current/`
- Check `{id}` and `current/` path casing.
- If MinIO returns `NoSuchKey`, the object path is wrong or not uploaded.

Wrong path prefix
- Expected path is `/games/<id>/current/...` (bucket is `games`).
- If MinIO returns `NoSuchBucket`, the proxy is stripping `/games` (no trailing slash on `proxy_pass`).

Wrong content-type (browser downloads file)
- Check response headers for `Content-Type` in devtools.
- Inspect object metadata: `mc stat local/games/{id}/current/index.html`
- Re-upload with correct `Content-Type` metadata if needed.

### Verification checklist
- `/games/{id}/current/index.html` returns 200
- JS/CSS/images load without 404s in browser devtools
- Response headers include `Cache-Control` (and `Expires` if configured)
- `Content-Type` matches file types (HTML, JS, CSS, images)
- `/api/health` still returns 200
