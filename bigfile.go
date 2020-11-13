package bigfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
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
	}
}

// Remove .
func Remove(dir string) error {
	return os.RemoveAll(dir)
}

// Close .
func (f *File) Close() {
	if f.fd > -1 {
		unix.Close(f.fd)
		f.fd = -1
	}
}

// Size .
func (f *File) Size() (int64, error) {
	d, err := os.Open(f.dir)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(f.dir, dirPerm); err != nil {
				return 0, fmt.Errorf("os.MkdirAll: %w", err)
			}
			return f.Size()
		}
		return 0, fmt.Errorf("os.Open: %w", err)
	}
	list, err := d.Readdirnames(-1)
	d.Close()
	if err != nil {
		return 0, fmt.Errorf("d.Readdir: %w", err)
	}

	fileCnt := int64(len(list))
	if fileCnt == 0 {
		return 0, nil
	}
	filledSize := int64(fileCnt-1) * f.partSize
	var stat unix.Stat_t
	if err := unix.Stat(filepath.Join(f.dir, padZeros(fileCnt-1)), &stat); err != nil {
		return 0, fmt.Errorf("unix.Stat: %w", err)
	}
	return stat.Size + filledSize, nil
}

// Seek .
func (f *File) Seek(offset int64) error {
	var err error
	err = f.move(offset)
	if err != nil {
		return fmt.Errorf("bigfile.move: %w", err)
	}
	partOff := offset % f.partSize
	offset, err = unix.Seek(f.fd, partOff, 0)
	return err
}

// Read .
func (f *File) Read(b []byte) error {
	var err error
	if f.fd < 0 || f.currentIndex < 0 {
		err = f.move(0)
		if err != nil {
			return fmt.Errorf("f.move: %w", err)
		}
	}

	var nextFileBytes []byte
	if int64(len(b))+(f.offset%f.partSize) > f.partSize {
		nextFileBytes = b[f.partSize-(f.offset%f.partSize):]
		b = b[:f.partSize-(f.offset%f.partSize)]
	}
	_, err = unix.Read(f.fd, b)
	if err != nil {
		return fmt.Errorf("unix.Read: %w", err)
	}
	if len(nextFileBytes) > 0 {
		err = f.move(f.offset + int64(len(b)))
		if err != nil {
			return fmt.Errorf("bigfile.move: %w", err)
		}
		err = f.Read(nextFileBytes)
		if err != nil {
			return fmt.Errorf("bigfile.Read: %w", err)
		}
	}

	return nil
}

// Write .
func (f *File) Write(b []byte) error {
	var err error
	if f.fd < 0 || f.currentIndex < 0 {
		err = f.move(0)
		if err != nil {
			return fmt.Errorf("f.move: %w", err)
		}
	}

	var nextFileBytes []byte
	if int64(len(b))+(f.offset%f.partSize) > f.partSize {
		nextFileBytes = b[f.partSize-(f.offset%f.partSize):]
		b = b[:f.partSize-(f.offset%f.partSize)]
	}
	_, err = unix.Write(f.fd, b)
	if err != nil {
		return fmt.Errorf("unix.Write: %w", err)
	}
	if len(nextFileBytes) > 0 {
		err = f.move(f.offset + int64(len(b)))
		if err != nil {
			return fmt.Errorf("bigfile.move: %w", err)
		}
		err = f.Write(nextFileBytes)
		if err != nil {
			return fmt.Errorf("bigfile.Write: %w", err)
		}
	}

	return nil
}

// ReadAt .
func (f *File) ReadAt(b []byte, off int64) error {
	var err error
	err = f.move(off)
	if err != nil {
		return fmt.Errorf("f.move %d: %w", off, err)
	}
	partOff := off % f.partSize
	var nextFileBytes []byte
	if int64(len(b))+partOff > f.partSize {
		nextFileBytes = b[f.partSize-partOff:]
		b = b[:f.partSize-partOff]
	}

	_, err = unix.Pread(f.fd, b, partOff)
	if err != nil {
		return fmt.Errorf("f.file.ReadAt file %d, off %d, part size %d: %w", f.currentIndex, off, f.partSize, err)
	}
	if len(nextFileBytes) > 0 {
		err = f.ReadAt(nextFileBytes, off+int64(len(b)))
		if err != nil {
			return fmt.Errorf("nextFileRead: %w", err)
		}
	}
	return nil
}

// WriteAt .
func (f *File) WriteAt(b []byte, off int64) error {
	var err error
	err = f.move(off)
	if err != nil {
		return fmt.Errorf("f.move %d: %w", off, err)
	}
	partOff := off % f.partSize
	var nextFileBytes []byte
	if int64(len(b))+partOff > f.partSize {
		nextFileBytes = b[f.partSize-partOff:]
		b = b[:f.partSize-partOff]
	}
	_, err = unix.Pwrite(f.fd, b, partOff)
	if err != nil {
		return fmt.Errorf("f.file.WriteAt file %d, off %d, part size %d: %w", f.currentIndex, partOff, f.partSize, err)
	}
	if len(nextFileBytes) > 0 {
		err = f.WriteAt(nextFileBytes, off+int64(len(b)))
		if err != nil {
			return fmt.Errorf("nextFileWrite: %w", err)
		}
	}
	return nil
}

func (f *File) move(off int64) error {
	f.offset = off
	var err error
	if f.fd < 0 || off/f.partSize != f.currentIndex {
		if f.fd >= 0 {
			unix.Close(f.fd)
		}
		f.currentIndex = off / f.partSize
		newPath := filepath.Join(f.dir, padZeros(f.currentIndex))
		f.fd, err = unix.Open(newPath, unix.O_RDWR, filePerm)
		if err != nil {
			if err == unix.ENOENT {
				err = os.MkdirAll(f.dir, dirPerm)
				if err != nil {
					return err
				}
				f.fd, err = unix.Open(newPath, unix.O_CREAT|unix.O_RDWR, filePerm)
				if err != nil {
					return err
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
