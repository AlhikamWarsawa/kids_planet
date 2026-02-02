package utils

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ZipInputError struct {
	Reason string
}

func (e ZipInputError) Error() string {
	if e.Reason == "" {
		return "invalid zip file"
	}
	return e.Reason
}

func SafeUnzip(reader io.ReaderAt, size int64, dest string) ([]string, error) {
	if reader == nil {
		return nil, errors.New("zip reader is nil")
	}
	if size < 0 {
		return nil, errors.New("zip size is invalid")
	}
	if strings.TrimSpace(dest) == "" {
		return nil, errors.New("destination is empty")
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return nil, fmt.Errorf("create destination: %w", err)
	}

	zr, err := zip.NewReader(reader, size)
	if err != nil {
		return nil, ZipInputError{Reason: "invalid zip file"}
	}

	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return nil, fmt.Errorf("resolve destination: %w", err)
	}

	extracted := make([]string, 0, len(zr.File))
	for _, f := range zr.File {
		name := strings.ReplaceAll(f.Name, "\\", "/")
		if err := validateZipEntryName(name); err != nil {
			return nil, ZipInputError{Reason: err.Error()}
		}
		if f.Mode()&os.ModeSymlink != 0 {
			return nil, ZipInputError{Reason: fmt.Sprintf("zip entry %q is a symlink", name)}
		}

		relPath := filepath.Clean(filepath.FromSlash(name))
		if relPath == "." {
			return nil, ZipInputError{Reason: fmt.Sprintf("zip entry %q has invalid path", name)}
		}

		targetPath := filepath.Join(destAbs, relPath)
		if !isWithinBase(destAbs, targetPath) {
			return nil, ZipInputError{Reason: fmt.Sprintf("zip entry %q is outside destination", name)}
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return nil, fmt.Errorf("create directory: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return nil, fmt.Errorf("create directory: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("open zip entry: %w", err)
		}
		out, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			_ = rc.Close()
			return nil, fmt.Errorf("create file: %w", err)
		}

		_, copyErr := io.Copy(out, rc)
		closeErr := out.Close()
		rcErr := rc.Close()
		if copyErr != nil {
			return nil, fmt.Errorf("write file: %w", copyErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("close file: %w", closeErr)
		}
		if rcErr != nil {
			return nil, fmt.Errorf("close zip entry: %w", rcErr)
		}

		extracted = append(extracted, filepath.ToSlash(relPath))
	}

	return extracted, nil
}

func validateZipEntryName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("zip entry has empty name")
	}
	if strings.ContainsRune(name, '\x00') {
		return fmt.Errorf("zip entry %q has invalid character", name)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("zip entry %q contains invalid path segment", name)
	}
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "\\") {
		return fmt.Errorf("zip entry %q is absolute", name)
	}
	if path.IsAbs(name) {
		return fmt.Errorf("zip entry %q is absolute", name)
	}
	if len(name) >= 2 && name[1] == ':' {
		return fmt.Errorf("zip entry %q is absolute", name)
	}
	if strings.HasPrefix(name, "//") || strings.HasPrefix(name, "\\\\") {
		return fmt.Errorf("zip entry %q is absolute", name)
	}
	return nil
}

func isWithinBase(base, target string) bool {
	base = filepath.Clean(base)
	target = filepath.Clean(target)
	if base == target {
		return true
	}
	sep := string(os.PathSeparator)
	if !strings.HasSuffix(base, sep) {
		base += sep
	}
	return strings.HasPrefix(target, base)
}
