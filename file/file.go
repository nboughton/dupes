package file

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nboughton/go-utils/fs"
	"gopkg.in/cheggaaa/pb.v1"
)

// File stores a copy of the md5 hash and paths to all files that
// match that hash
type File struct {
	Hash  string
	Paths []string
}

// New returns a new File with its hash and first found path
func New(hash, path string) *File {
	return &File{
		Hash:  hash,
		Paths: []string{path},
	}
}

// AddPath adds a matching path to an existing File record
func (f *File) AddPath(path string) {
	f.Paths = append(f.Paths, path)
}

// Keep removes the files at all paths except for the 1 specifed by its index number
func (f *File) Keep(idx int) error {
	if idx > len(f.Paths) {
		return fmt.Errorf("Error: Invalid index: [%d]", idx)
	}

	for i := range f.Paths {
		if i != idx {
			if err := os.Remove(f.Paths[i]); err != nil {
				return err
			}
		}
	}

	return nil
}

// Index returns the index strings of all paths for file
func (f *File) Index() (idx []string) {
	for i, p := range f.Paths {
		idx = append(idx, fmt.Sprintf("[%d] %s", i, p))
	}

	return idx
}

// Valid tests whether a file is the kind that we want to test
func Valid(path string) bool {
	if s, err := fs.IsSymlink(path); err != nil || s {
		return false
	}

	f, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !f.IsDir() && f.Size() > 0 && f.Size() < 5e+8
}

// Hash returns the sha256 hash of a given file
func Hash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return string(h.Sum(nil)), nil
}

// ReadTree recursively reads dir and produces a map of hash->File relationships
func ReadTree(dir string, ignoreDot bool) (map[string]*File, []error) {
	h, errors, bar := make(map[string]*File), []error{}, pb.StartNew(Count(dir, ignoreDot))

	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err == nil && Valid(path) {
			if ignoreDot && strings.HasPrefix(f.Name(), ".") {
				return nil
			}

			sum, err := Hash(path)
			if err != nil {
				errors = append(errors, err)
				fmt.Println(err)
				bar.Increment()
				return nil
			}

			// Create or append record
			if _, ok := h[sum]; !ok {
				h[sum] = New(sum, path)
			} else {
				h[sum].AddPath(path)
			}

			bar.Increment()
			time.Sleep(time.Millisecond)
		}

		return nil
	})
	bar.Finish()

	return h, errors
}

// Count traverses dir and returns the number of files that are valid for checking
func Count(dir string, ignoreDot bool) int {
	c := 0

	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err == nil && Valid(path) {
			if ignoreDot && strings.HasPrefix(f.Name(), ".") {
				return nil
			}

			c++
		}

		return nil
	})

	return c
}
