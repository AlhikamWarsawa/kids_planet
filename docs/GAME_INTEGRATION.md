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