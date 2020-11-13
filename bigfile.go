package bigfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"syscall"
)

const filePerm = 0600
const dirPerm = 0700

// File .
type File struct {
	partSize     int64
	dir          string
	fd           int
	currentIndex int64
	offset       int64
}

// Open .
func Open(dir string, partSize int64) *File {
	return &File{
		partSize:     partSize,
		dir:          dir,
		currentIndex: -1,
		fd:           -1,
		offset:       0,
	}
}

// Remove .
func Remove(dir string) error {
	return os.RemoveAll(dir)
}

// Close .
func (f *File) Close() error {
	if f.fd > -1 {
		f.fd = -1
		return syscall.Close(f.fd)
	}
	return nil
}

// Size .
func (f *File) Size() (int64, error) {
	var stat syscall.Stat_t
	if err := syscall.Stat(f.dir, &stat); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err = os.MkdirAll(f.dir, dirPerm); err != nil {
				return 0, fmt.Errorf("os.MkdirAll: %w", err)
			}
			return f.Size()
		}
		return 0, fmt.Errorf("syscall.Stat: %w", err)
	}
	fileCnt := (stat.Size - 64) / 32
	if fileCnt == 0 {
		return 0, nil
	}
	filledSize := int64(fileCnt-1) * f.partSize

	if err := syscall.Stat(filepath.Join(f.dir, padZeros(fileCnt-1)), &stat); err != nil {
		return 0, fmt.Errorf("syscall.Stat: %w", err)
	}
	return stat.Size + filledSize, nil
}

// Seek .
func (f *File) Seek(offset int64, whence int) (int64, error) {
	var err error
	err = f.move(offset)
	if err != nil {
		return 0, fmt.Errorf("bigfile.move: %w", err)
	}
	partOff := offset % f.partSize
	offset, err = syscall.Seek(f.fd, partOff, whence)
	return offset + f.partSize*f.currentIndex, err
}

// Read .
func (f *File) Read(b []byte) (int, error) {
	var err error
	if f.fd < 0 || f.currentIndex < 0 {
		err = f.move(0)
		if err != nil {
			return 0, fmt.Errorf("bigfile.move: %w", err)
		}
		_, err = syscall.Seek(f.fd, f.offset%f.partSize, 0)
		if err != nil {
			return 0, fmt.Errorf("syscall.Seek: %w", err)
		}
	}
	var nextFileBytes []byte
	if int64(len(b))+(f.offset%f.partSize) > f.partSize {
		nextFileBytes = b[f.partSize-(f.offset%f.partSize):]
		b = b[:f.partSize-(f.offset%f.partSize)]
	}
	var n int
	n, err = syscall.Read(f.fd, b)
	if len(nextFileBytes) > 0 {
		err = f.move(f.offset + int64(n))
		if err != nil {
			return 0, fmt.Errorf("bigfile.move: %w", err)
		}
		var n2 int
		n2, err = f.Read(nextFileBytes)
		if err != nil {
			return 0, fmt.Errorf("bigfile.Read: %w", err)
		}
		n += n2
	}
	return n, nil
}

// Write .
func (f *File) Write(b []byte) (int, error) {
	var err error
	if f.fd < 0 || f.currentIndex < 0 {
		err = f.move(0)
		if err != nil {
			return 0, fmt.Errorf("bigfile.move: %w", err)
		}
		_, err = syscall.Seek(f.fd, 0, 0)
		if err != nil {
			return 0, fmt.Errorf("syscall.Seek: %w", err)
		}
	}
	var nextFileBytes []byte
	if int64(len(b))+(f.offset%f.partSize) > f.partSize {
		nextFileBytes = b[f.partSize-(f.offset%f.partSize):]
		b = b[:f.partSize-(f.offset%f.partSize)]
	}
	var n int
	n, err = syscall.Write(f.fd, b)
	if err != nil {
		return 0, err
	}
	if len(nextFileBytes) > 0 {
		err = f.move((f.currentIndex + 1) * f.partSize)
		if err != nil {
			return 0, fmt.Errorf("bigfile.move: %w", err)
		}
		var n2 int
		n2, err = f.Write(nextFileBytes)
		if err != nil {
			return 0, fmt.Errorf("bigfile.Write: %w", err)
		}
		n += n2
	}
	return n, nil
}

// ReadAt .
func (f *File) ReadAt(b []byte, off int64) (int, error) {
	var err error
	err = f.move(off)
	if err != nil {
		return 0, fmt.Errorf("bigfile.move: %w", err)
	}
	partOff := off % f.partSize
	var nextFileBytes []byte
	if int64(len(b))+partOff > f.partSize {
		nextFileBytes = b[f.partSize-partOff:]
		b = b[:f.partSize-partOff]
	}
	var n int
	n, err = syscall.Pread(f.fd, b, partOff)
	if err != nil {
		return 0, fmt.Errorf("syscall.Pread: %w", err)
	}
	if len(nextFileBytes) > 0 {
		var n2 int
		n2, err = f.ReadAt(nextFileBytes, off+int64(len(b)))
		if err != nil {
			return 0, fmt.Errorf("bigfile.ReadAt: %w", err)
		}
		n += n2
	}
	return n, nil
}

// WriteAt .
func (f *File) WriteAt(b []byte, off int64) (int, error) {
	var err error
	err = f.move(off)
	if err != nil {
		return 0, fmt.Errorf("bigfile.move: %w", err)
	}
	partOff := off % f.partSize
	var nextFileBytes []byte
	if int64(len(b))+partOff > f.partSize {
		nextFileBytes = b[f.partSize-partOff:]
		b = b[:f.partSize-partOff]
	}
	var n int
	n, err = syscall.Pwrite(f.fd, b, partOff)
	if err != nil {
		return 0, fmt.Errorf("syscall.Pwrite: %w", err)
	}
	if len(nextFileBytes) > 0 {
		var n2 int
		n2, err = f.WriteAt(nextFileBytes, off+int64(len(b)))
		if err != nil {
			return 0, fmt.Errorf("bigfile.WriteAt: %w", err)
		}
		n += n2
	}
	return n, nil
}

// WriteAt .
func (f *File) move(off int64) error {
	var err error
	f.offset = off
	if f.fd < 0 || off/f.partSize != f.currentIndex {
		if f.fd >= 0 {
			err = syscall.Close(f.fd)
			if err != nil {
				return fmt.Errorf("syscall.Close: %w", err)
			}
		}
		f.currentIndex = off / f.partSize
		newPath := filepath.Join(f.dir, padZeros(f.currentIndex))
		f.fd, err = syscall.Open(newPath, syscall.O_RDWR, filePerm)
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(f.dir, dirPerm)
				if err != nil {
					return fmt.Errorf("os.MkdirAll: %w", err)
				}
				f.fd, err = syscall.Open(newPath, syscall.O_CREAT|syscall.O_RDWR, filePerm)
				if err != nil {
					return fmt.Errorf("syscall.Open: %w", err)
				}
			}
		}
	}
	return err
}

func padZeros(i int64) string {
	var b strings.Builder
	is := strconv.FormatInt(i, 10)
	for b.Len()+len(is) < 12 {
		b.WriteRune('0')
	}
	b.WriteString(is)
	return b.String()
}
