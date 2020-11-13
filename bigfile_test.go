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
	dir := "TestSingleFileUse"
	f := bigfile.Open(dir, 10)
	defer f.Close()
	control := []byte("foo")
	_, err := f.WriteAt(control, 0)
	check(t, err)
	test := make([]byte, len(control))
	_, err = f.ReadAt(test, 0)
	check(t, err)
	if bytes.Compare(control, test) != 0 {
		t.Fatalf("%b != %b", control, test)
	}
	check(t, bigfile.Remove(dir))
}

func TestReadFromNotFirstFile(t *testing.T) {
	dir := "TestReadFromNotFirstFile"
	f := bigfile.Open(dir, 6)
	defer f.Close()
	control := []byte("foobarbaz")
	_, err := f.WriteAt(control, 0)
	check(t, err)
	test := make([]byte, 3)
	_, err = f.ReadAt(test, 6)
	check(t, err)
	if string(test) != "baz" {
		t.Fatalf("%s != %s", string(test), "baz")
	}
	check(t, bigfile.Remove(dir))
}

func TestWriteManyFiles(t *testing.T) {
	dir := "TestWriteManyFiles"
	f := bigfile.Open(dir, 3)
	defer f.Close()
	control := []byte("foobarbaz")
	_, err := f.WriteAt(control, 0)
	check(t, err)
	test := make([]byte, 3)
	_, err = f.ReadAt(test, 6)
	check(t, err)
	if string(test) != "baz" {
		t.Fatalf("%s != %s", string(test), "baz")
	}
	check(t, bigfile.Remove(dir))
}

func TestReadCrossManyFiles(t *testing.T) {
	dir := "TestReadCrossManyFiles"
	f := bigfile.Open(dir, 3)
	defer f.Close()
	control := []byte("foobarbaz")
	_, err := f.WriteAt(control, 0)
	check(t, err)
	test := make([]byte, len(control))
	_, err = f.ReadAt(test, 0)
	check(t, err)
	if bytes.Compare(control, test) != 0 {
		t.Fatalf("%b != %b", control, test)
	}
	check(t, bigfile.Remove(dir))
}

func TestSeekRead(t *testing.T) {
	dir := "TestSeekRead"
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
	_, err = f.Seek(3, 0)
	check(t, err)
	test := make([]byte, 6)
	n, err = f.Read(test)
	check(t, err)
	if n != len(test) {
		t.Fatalf("n %d != %d", n, len(test))
	}
	if bytes.Compare(control[3:], test) != 0 {
		t.Fatalf("%b != %b", control[3:], test)
	}
	check(t, bigfile.Remove(dir))
}

func TestSize(t *testing.T) {
	dir := "TestSize"
	f := bigfile.Open(dir, 5)
	defer f.Close()
	control := []byte("foobarbaz")
	_, err := f.WriteAt(control, 0)
	check(t, err)
	s, err := f.Size()
	check(t, err)
	if s != 9 {
		t.Fatalf("size %d != %d", s, 9)
	}
	check(t, bigfile.Remove(dir))
}

func TestZeroSize(t *testing.T) {
	dir := "TestZeroSize"
	f := bigfile.Open(dir, 5)
	defer f.Close()
	s, err := f.Size()
	check(t, err)
	if s != 0 {
		t.Fatalf("size %d != %d", s, 0)
	}
	check(t, bigfile.Remove(dir))
}
