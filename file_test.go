package safeio

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenFile_ok(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "safeio")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	name := filepath.Join(tmpDir, "test1.txt")
	mode := os.FileMode(0644)

	f, err := OpenFile(name, mode)
	require.NoError(t, err)
	defer f.Close() // to prevent leaks

	data := []byte("hello world\n")

	n, err := f.Write(data)
	require.NoError(t, err)
	require.Equal(t, 12, n)

	err = f.Commit()
	require.NoError(t, err)

	err = f.Close()
	require.NoError(t, err)

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

func TestOpenFile_elective_abort(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "safeio")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	name := filepath.Join(tmpDir, "test1.txt")
	mode := os.FileMode(0644)

	f, err := OpenFile(name, mode)
	require.NoError(t, err)
	defer f.Close() // to prevent leaks

	data := []byte("hello world\n")

	n, err := f.Write(data)
	require.NoError(t, err)
	require.Equal(t, 12, n)

	// no commit

	err = f.Close()
	require.NoError(t, err)

	// verify

	_, err = ioutil.ReadFile(name)
	require.True(t, os.IsNotExist(err))

	// no files

	list, err := ioutil.ReadDir(tmpDir)
	require.NoError(t, err)
	require.Equal(t, 0, len(list))
}

var testErrDiskBroke = errors.New("disk broke!")

func TestOpenFile_writeErrorOnCommit_abort(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "safeio")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	name := filepath.Join(tmpDir, "test1.txt")
	mode := os.FileMode(0644)

	f, err := OpenFile(name, mode)
	require.NoError(t, err)
	defer f.Close() // to prevent leaks

	data := []byte("hello world\n")

	n, err := f.Write(data)
	require.NoError(t, err)
	require.Equal(t, 12, n)

	// simulate disk wonkiness
	f.setErr(testErrDiskBroke)

	err = f.Commit()
	require.Equal(t, testErrDiskBroke, err)

	err = f.Close()
	require.Equal(t, testErrDiskBroke, err)

	// verify

	_, err = ioutil.ReadFile(name)
	require.True(t, os.IsNotExist(err))

	// no files

	list, err := ioutil.ReadDir(tmpDir)
	require.NoError(t, err)
	require.Equal(t, 0, len(list))
}

func TestOpenFile_writeErrorOnWrite_abort(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "safeio")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	name := filepath.Join(tmpDir, "test1.txt")
	mode := os.FileMode(0644)

	f, err := OpenFile(name, mode)
	require.NoError(t, err)
	defer f.Close() // to prevent leaks

	data := []byte("hello world\n")

	// write half before the error
	n, err := f.Write(data[0:6])
	require.NoError(t, err)
	require.Equal(t, 6, n)

	// simulate disk wonkiness
	f.setErr(testErrDiskBroke)

	// and try to write half after the error
	n, err = f.Write(data[6:])
	require.Equal(t, testErrDiskBroke, err)
	require.Equal(t, 0, n)

	err = f.Commit()
	require.Equal(t, testErrDiskBroke, err)

	err = f.Close()
	require.Equal(t, testErrDiskBroke, err)

	// verify

	_, err = ioutil.ReadFile(name)
	require.True(t, os.IsNotExist(err))

	// no files

	list, err := ioutil.ReadDir(tmpDir)
	require.NoError(t, err)
	require.Equal(t, 0, len(list))
}
