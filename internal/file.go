package gopher

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// closeWithErr preserves an earlier error, otherwise returning a close error.
func closeWithErr(closer io.Closer, err *error) {
	if closeErr := closer.Close(); closeErr != nil && *err == nil {
		*err = fmt.Errorf("close: %w", closeErr)
	}
}

// readRootedFile opens only the final path component within its parent.
func readRootedFile(path string) (data []byte, err error) {
	file, err := openRootedFile(path)
	if err != nil {
		return nil, err
	}
	defer closeWithErr(file, &err)
	return io.ReadAll(file)
}

func openRootedFile(path string) (*os.File, error) {
	cleanPath := filepath.Clean(path)
	resolvedPath, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		return nil, err
	}
	return os.OpenInRoot(filepath.Dir(resolvedPath), filepath.Base(resolvedPath))
}

func writablePath(path string) (string, error) {
	cleanPath := filepath.Clean(path)
	resolvedPath, err := filepath.EvalSymlinks(cleanPath)
	if err == nil {
		return resolvedPath, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	linkTarget, linkErr := os.Readlink(cleanPath)
	if linkErr == nil {
		if !filepath.IsAbs(linkTarget) {
			linkTarget = filepath.Join(filepath.Dir(cleanPath), linkTarget)
		}
		return writablePath(linkTarget)
	}
	resolvedDir, err := filepath.EvalSymlinks(filepath.Dir(cleanPath))
	if err != nil {
		return "", errors.Join(linkErr, err)
	}
	return filepath.Join(resolvedDir, filepath.Base(cleanPath)), nil
}
