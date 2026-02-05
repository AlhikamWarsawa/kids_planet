# Game ZIP Integration

This document describes the required structure for game ZIP uploads.

## Required structure
- `index.html` must exist at the ZIP root.
- All paths must be relative (no leading `/`, `\`, or drive letters).
- No `..` segments anywhere in any entry name.
- Symlinks are not allowed.
- Nested assets are allowed (e.g., `assets/`, `js/`, `css/`).

## Extraction behavior
- The ZIP is extracted to a temporary directory with zip-slip protections.
- Extracted files are uploaded to `games/{id}/current/{relative_path}`.
- If `index.html` is missing at the root, the upload is rejected.

## Game Integration Guideline
- The ZIP must contain `index.html` at the root (no nested folder).
- The playable URL is `/games/{id}/current/index.html`.
- Common errors:
- `INVALID_ZIP`: The file is not a valid ZIP or contains unsafe paths.
- `ZIP_TOO_LARGE`: The ZIP exceeds the upload size limit.
- `MISSING_INDEX_HTML`: `index.html` was not found at the ZIP root.
- `INTERNAL_ERROR`: The server failed while processing the ZIP.
- If an upload fails, confirm:
- The ZIP opens locally without errors.
- `index.html` is at the top level.
- There are no absolute paths, symlinks, or `..` segments.

## Example ZIP layout
```
index.html
assets/
assets/logo.png
css/
css/styles.css
js/
js/game.js
```
