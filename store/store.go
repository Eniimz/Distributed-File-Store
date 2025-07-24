package store

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
)

// reading, writing to a store

type PathKey struct {
	PathName string
	FileName string
}

type pathTransFormFunc func(string) PathKey

type StoreOpts struct {
	pathTransFormFunc pathTransFormFunc
}

type Store struct {
	StoreOpts
}

func (p PathKey) firstPath(path string) string {

	if len(path) == 0 {
		fmt.Printf("No path to Split")
		return "No Path..."
	}

	return strings.Split(path, "/")[0]

}

func (p PathKey) fullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.FileName)
}

func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

func Has(key string) bool {

	pathKey := CASPathTransformFunc(key)

	_, err := os.Stat(pathKey.fullPath())
	if err == fs.ErrNotExist {
		return false
	}

	return true

}

func CASPathTransformFunc(key string) PathKey {

	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	strLen := len(hashStr) / blockSize

	//make a slice of strings thats initial len of 8
	paths := make([]string, strLen)
	for i := 0; i < strLen; i++ {

		//start = initial val * len of block,
		//end = end of first 8 len str + len of block
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		FileName: hashStr,
	}
}

func (s *Store) writeStream(key string, r io.Reader) error {

	//hashing
	pathKey := CASPathTransformFunc(key) // sdad/sds/sds.../sfwd

	if err := os.MkdirAll(pathKey.PathName, os.ModePerm); err != nil {
		return err
	}
	//we make a PathKey struct, so we can easily access the hashStr without /
	fullPath := pathKey.fullPath()

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}

	//f is the returned file that is created
	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	fmt.Printf("Written %d bytes to the disk %s", n, fullPath)

	return nil
}

func (s *Store) Read(key string) (io.Reader, error) {

	f, err := s.ReadStream(key)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)

	return buf, err

}

// to give user more authority on what to do with the read data,
// how and when to read or close it
func (s *Store) ReadStream(key string) (io.ReadCloser, error) {

	pathKey := CASPathTransformFunc(key)
	// file, err := os.Open(pathKey.fullPath())
	return os.Open(pathKey.fullPath())
}

func (s *Store) Delete(key string) error {

	pathKey := CASPathTransformFunc(key)

	defer func() {
		fmt.Printf("deleted [%s] from disk", pathKey.FileName)
	}()

	fmt.Printf("The first Path: %s\n", pathKey.firstPath(pathKey.fullPath()))

	return os.RemoveAll(pathKey.firstPath(pathKey.fullPath()))

}
