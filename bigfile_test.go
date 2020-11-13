package bigfile_test

import (
	"bytes"
	"testing"

	"github.com/twharmon/bigfile"
)

func check(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestSingleFileUse(t *testing.T) {
	dir := "/tmp/TestSingleFileUse"
	f := bigfile.Open(dir, 10)
	defer f.Close()
	control := []byte("foo")
	var err error
	var n int
	n, err = f.WriteAt(control, 0)
	check(t, err)
	if n != len(control) {
		t.Fatalf("n %d != %d", n, len(control))
	}
	test := make([]byte, len(control))
	n, err = f.ReadAt(test, 0)
	check(t, err)
	if n != len(test) {
		t.Fatalf("n %d != %d", n, len(test))
	}
	if bytes.Compare(control, test) != 0 {
		t.Fatalf("%b != %b", control, test)
	}
	check(t, bigfile.Remove(dir))
}

func TestReadFromNotFirstFile(t *testing.T) {
	dir := "/tmp/TestReadFromNotFirstFile"
	f := bigfile.Open(dir, 6)
	defer f.Close()
	control := []byte("foobarbaz")
	var err error
	var n int
	n, err = f.WriteAt(control, 0)
	check(t, err)
	if n != len(control) {
		t.Fatalf("n %d != %d", n, len(control))
	}
	test := make([]byte, 3)
	n, err = f.ReadAt(test, 6)
	if n != len(test) {
		t.Fatalf("n %d != %d", n, len(test))
	}
	check(t, err)
	if string(test) != "baz" {
		t.Fatalf("%s != %s", string(test), "baz")
	}
	check(t, bigfile.Remove(dir))
}

func TestWriteManyFiles(t *testing.T) {
	dir := "/tmp/TestWriteManyFiles"
	f := bigfile.Open(dir, 3)
	defer f.Close()
	control := []byte("foobarbaz")
	var err error
	var n int
	n, err = f.WriteAt(control, 0)
	check(t, err)
	if n != len(control) {
		t.Fatalf("n %d != %d", n, len(control))
	}
	test := make([]byte, 3)
	n, err = f.ReadAt(test, 6)
	check(t, err)
	if n != len(test) {
		t.Fatalf("n %d != %d", n, len(test))
	}
	if string(test) != "baz" {
		t.Fatalf("%s != %s", string(test), "baz")
	}
	check(t, bigfile.Remove(dir))
}

func TestReadCrossManyFiles(t *testing.T) {
	dir := "/tmp/TestReadCrossManyFiles"
	f := bigfile.Open(dir, 3)
	defer f.Close()
	control := []byte("foobarbaz")
	var err error
	var n int
	n, err = f.WriteAt(control, 0)
	check(t, err)
	if n != len(control) {
		t.Fatalf("n %d != %d", n, len(control))
	}
	test := make([]byte, len(control))
	n, err = f.ReadAt(test, 0)
	check(t, err)
	if n != len(test) {
		t.Fatalf("n %d != %d", n, len(test))
	}
	if bytes.Compare(control, test) != 0 {
		t.Fatalf("%b != %b", control, test)
	}
	check(t, bigfile.Remove(dir))
}

func TestSeekRead(t *testing.T) {
	dir := "/tmp/TestSeekRead"
	f := bigfile.Open(dir, 3)
	defer f.Close()

	control := []byte("foobarbaz")
	var n int
	var err error
	n, err = f.Write(control)
	check(t, err)
	if n != len(control) {
		t.Fatalf("n %d != %d", n, len(control))
	}
	var off int64
	off, err = f.Seek(4, 0)
	check(t, err)
	if off != 4 {
		t.Fatalf("off %d != %d", off, 4)
	}
	test := make([]byte, 5)
	n, err = f.Read(test)
	check(t, err)
	if n != len(test) {
		t.Fatalf("n %d != %d", n, len(test))
	}
	if bytes.Compare(control[4:], test) != 0 {
		t.Fatalf("%b != %b", control[4:], test)
	}
	check(t, bigfile.Remove(dir))
}

func TestReadFirst(t *testing.T) {
	dir := "/tmp/TestReadFirst"
	f := bigfile.Open(dir, 3)
	defer f.Close()
	control := []byte("foobarbaz")
	var n int
	var err error
	n, err = f.Write(control)
	check(t, err)
	if n != len(control) {
		t.Fatalf("n %d != %d", n, len(control))
	}
	f.Close()

	f2 := bigfile.Open(dir, 3)
	defer f2.Close()

	test := make([]byte, 7)
	n, err = f.Read(test)
	check(t, err)
	if n != len(test) {
		t.Fatalf("n %d != %d", n, len(test))
	}
	if bytes.Compare(control[:7], test) != 0 {
		t.Fatalf("%b != %b", control[:7], test)
	}
	check(t, bigfile.Remove(dir))
}

func TestSize(t *testing.T) {
	dir := "/tmp/TestSize"
	f := bigfile.Open(dir, 5)
	defer f.Close()
	control := []byte("foobarbaz")
	var n int
	var err error
	n, err = f.WriteAt(control, 0)
	check(t, err)
	if n != len(control) {
		t.Fatalf("n %d != %d", n, len(control))
	}
	s, err := f.Size()
	check(t, err)
	if s != 9 {
		t.Fatalf("size %d != %d", s, 9)
	}
	check(t, bigfile.Remove(dir))
}

func TestZeroSize(t *testing.T) {
	dir := "/tmp/TestZeroSize"
	f := bigfile.Open(dir, 5)
	defer f.Close()
	s, err := f.Size()
	check(t, err)
	if s != 0 {
		t.Fatalf("size %d != %d", s, 0)
	}
	check(t, bigfile.Remove(dir))
}
