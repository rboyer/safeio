package safeio

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteToFile_ok(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "safeio")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	name := filepath.Join(tmpDir, "test1.txt")
	mode := os.FileMode(0644)

	data := []byte("hello world\n")

	n, err := WriteToFile(bytes.NewReader(data), name, mode)
	require.NoError(t, err)
	require.Equal(t, int64(len(data)), n)

	// verify

	read, err := ioutil.ReadFile(name)
	require.NoError(t, err)
	require.Equal(t, data, read)

	fi, err := os.Stat(name)
	require.NoError(t, err)
	require.Equal(t, mode, fi.Mode().Perm())

	// this was the only file

	list, err := ioutil.ReadDir(tmpDir)
	require.NoError(t, err)
	require.Equal(t, 1, len(list))
}

func TestWriteToFile_sourceFails(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "safeio")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	name := filepath.Join(tmpDir, "test1.txt")
	mode := os.FileMode(0644)

	// It's hard to simulate disk wonkiness here, but we can test most of this
	// by simulating SOURCE wonkiness.
	data := []byte("hello world\n")
	fr := &failingReader{Reader: bytes.NewReader(data), failAfterBytes: 5}

	_, err = WriteToFile(fr, name, mode)
	require.Equal(t, testErrDiskBroke, err)

	// verify
	_, err = ioutil.ReadFile(name)
	require.True(t, os.IsNotExist(err))

	// no files

	list, err := ioutil.ReadDir(tmpDir)
	require.NoError(t, err)
	require.Equal(t, 0, len(list))
}

type failingReader struct {
	io.Reader
	failAfterBytes int
}

func (r *failingReader) Read(p []byte) (n int, err error) {
	if r.failAfterBytes <= 0 {
		return 0, testErrDiskBroke
	}

	n, err = r.Reader.Read(p)
	if err != nil {
		return n, err
	}

	r.failAfterBytes -= n
	if r.failAfterBytes < 0 {
		return n, testErrDiskBroke
	}

	return n, nil
}
