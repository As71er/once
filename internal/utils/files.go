package utils

import (
	"fmt"
	"hash/crc64"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type FileManager interface {
	CreateDir(name string) (string, error)
	DeleteDirRecursive(path string) error
	DeleteFile(name, dir string) error
	WriteFile(file io.ReadCloser, name, dir string) (string, error)
	FileExists(name, dir string) (bool, error)
	GetUploadDir(name string) string
	GetFileWriter(name, dir string) (io.WriteCloser, string, error)
	OpenFile(name, dir string) (*os.File, os.FileInfo, error)
	RenameFile(oldName, newName, dir string) (string, error)
}

type LocalArchive struct {
	UploadDir string
}

func NewLocalArchive(dir string) *LocalArchive {
	return &LocalArchive{
		UploadDir: dir,
	}
}

func (la *LocalArchive) CreateDir(name string) (string, error) {
	path := filepath.Join(la.UploadDir, name)

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return "", err
	}

	return path, nil
}

func (la *LocalArchive) DeleteDirRecursive(dir string) error {
	path := filepath.Join(la.UploadDir, filepath.Base(dir))

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to delete %#q: %w", path, err)
	}

	return nil
}

func (la *LocalArchive) DeleteFile(name, dir string) error {
	uploadDir := filepath.Join(la.UploadDir, filepath.Base(dir))
	path := filepath.Join(uploadDir, filepath.Base(name))

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete %#q: %w", path, err)
	}

	return nil
}

func (la *LocalArchive) WriteFile(file io.ReadCloser, name, dir string) (string, error) {
	path := filepath.Join(dir, filepath.Base(name))

	dest, err := os.Create(path)
	if err != nil {
		slog.Error("failed to create destination file", slog.Any("error", err))
		return "", err
	}

	_, copyErr := io.Copy(dest, file)
	closeErr := dest.Close()

	if copyErr != nil {
		os.Remove(path)
		return "", copyErr
	}
	if closeErr != nil {
		os.Remove(path)
		return "", closeErr
	}

	return dest.Name(), nil
}

func (la *LocalArchive) GetFileWriter(name, dir string) (io.WriteCloser, string, error) {
	uploadDir := filepath.Join(la.UploadDir, filepath.Base(dir))

	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return nil, "", fmt.Errorf("failed to create upload directory %q: %w", uploadDir, err)
	}

	path := filepath.Join(uploadDir, filepath.Base(name))

	file, err := os.Create(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create file %q: %w", path, err)
	}

	return file, path, nil
}

func (la *LocalArchive) FileExists(name, dir string) (bool, error) {
	uploadDir := filepath.Join(la.UploadDir, filepath.Base(dir))
	path := filepath.Join(uploadDir, filepath.Base(name))

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if info.IsDir() {
		return false, fmt.Errorf("%s is a directory, not a file", path)
	}

	return true, nil
}

func (la *LocalArchive) RenameFile(oldName, newName, dir string) (string, error) {
	uploadDir := filepath.Join(la.UploadDir, filepath.Base(dir))
	oldPath := filepath.Join(uploadDir, filepath.Base(oldName))
	newPath := filepath.Join(uploadDir, filepath.Base(newName))

	info, err := os.Stat(oldPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("source file %q does not exist", oldPath)
		}
		return "", fmt.Errorf("failed to stat source file %q: %w", oldPath, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%q is a directory, not a file", oldPath)
	}

	if _, err := os.Stat(newPath); err == nil {
		return newPath, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to stat destination file %q: %w", newPath, err)
	}

	if oldPath == newPath {
		return newPath, nil
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return "", fmt.Errorf("failed to rename %q to %q: %w", oldPath, newPath, err)
	}

	return newPath, nil
}

func (la *LocalArchive) OpenFile(name, dir string) (*os.File, os.FileInfo, error) {
	uploadDir := filepath.Join(la.UploadDir, filepath.Base(dir))
	path := filepath.Join(uploadDir, filepath.Base(name))

	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file %q: %w", path, err)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, fmt.Errorf("failed to stat file %q: %w", path, err)
	}

	return file, info, nil
}

func (la *LocalArchive) GetUploadDir(name string) string {
	name = filepath.Base(name)

	return filepath.Join(la.UploadDir, name)
}

var crcTable = crc64.MakeTable(crc64.ECMA)

func GenerateChecksum(value string) string {
	now := time.Now()
	micros := now.UnixNano() / 1000

	data := fmt.Sprintf("%s:%d", value, micros)
	checksum := crc64.Checksum([]byte(data), crcTable)

	return fmt.Sprintf("%016x", checksum)
}

func GenerateChecksumShort(value string) string {
	now := time.Now()
	micros := now.UnixNano() / 1000

	data := fmt.Sprintf("%s:%d", value, micros)
	checksum := crc64.Checksum([]byte(data), crcTable)

	return fmt.Sprintf("%016x", checksum)[:8]
}
