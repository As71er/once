package utils

import (
	"errors"
	"fmt"
	"hash/crc64"
	"io"
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

var ErrFile = errors.New("file manager error")

func (la *LocalArchive) CreateDir(name string) (string, error) {
	path := filepath.Join(la.UploadDir, name)

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return "", fmt.Errorf("create dir %q: %w", path, errors.Join(ErrFile, err))
	}

	return path, nil
}

func (la *LocalArchive) DeleteDirRecursive(dir string) error {
	path := filepath.Join(la.UploadDir, filepath.Base(dir))

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("create dir %q: %w", path, errors.Join(ErrFile, err))
	}

	return nil
}

func (la *LocalArchive) DeleteFile(name, dir string) error {
	uploadDir := filepath.Join(la.UploadDir, filepath.Base(dir))
	path := filepath.Join(uploadDir, filepath.Base(name))

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete file %q: %w", path, errors.Join(ErrFile, err))
	}

	return nil
}

func (la *LocalArchive) WriteFile(file io.ReadCloser, name, dir string) (string, error) {
	path := filepath.Join(dir, filepath.Base(name))

	dest, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create file %q: %w", path, errors.Join(ErrFile, err))
	}

	_, copyErr := io.Copy(dest, file)
	closeErr := dest.Close()

	if copyErr != nil {
		os.Remove(path)
		return "", fmt.Errorf("copy data to %q: %w", path, ErrFile)
	}
	if closeErr != nil {
		os.Remove(path)
		return "", fmt.Errorf("close file %q: %w", path, errors.Join(ErrFile, err))
	}

	return dest.Name(), nil
}

func (la *LocalArchive) GetFileWriter(name, dir string) (io.WriteCloser, string, error) {
	uploadDir := filepath.Join(la.UploadDir, filepath.Base(dir))

	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return nil, "", fmt.Errorf("create upload dir %q: %w", uploadDir, errors.Join(ErrFile, err))
	}

	path := filepath.Join(uploadDir, filepath.Base(name))

	file, err := os.Create(path)
	if err != nil {
		return nil, "", fmt.Errorf("create file %q: %w", path, errors.Join(ErrFile, err))
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
		return false, fmt.Errorf("stat file %q: %w", path, errors.Join(ErrFile, err))
	}

	if info.IsDir() {
		return false, fmt.Errorf("check file %q: is a directory: %w", path, ErrFile)
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
			return "", fmt.Errorf("rename, file %#q does not exist: %w", oldPath, ErrFile)
		}
		return "", fmt.Errorf("stat file %#q: %w", oldPath, errors.Join(ErrFile, err))
	}
	if info.IsDir() {
		return "", fmt.Errorf("rename, %#q is a directory, not a file: %w", oldPath, ErrFile)
	}

	if _, err := os.Stat(newPath); err == nil {
		return newPath, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat file %q: %w", newPath, errors.Join(ErrFile, err))
	}

	if oldPath == newPath {
		return newPath, nil
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return "", fmt.Errorf("rename %q to %q: %w", oldPath, newPath, errors.Join(ErrFile, err))
	}

	return newPath, nil
}

func (la *LocalArchive) OpenFile(name, dir string) (*os.File, os.FileInfo, error) {
	uploadDir := filepath.Join(la.UploadDir, filepath.Base(dir))
	path := filepath.Join(uploadDir, filepath.Base(name))

	file, err := os.Open(path)

	if err != nil {
		return nil, nil, fmt.Errorf("open file %q: %w", path, errors.Join(ErrFile, err))
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, fmt.Errorf("stat file %q: %w", path, errors.Join(ErrFile, err))
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
