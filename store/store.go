package store

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/eniimz/cas/encryption"
)

// reading, writing to a store

const DefaultRootName = "Saga"

type PathKey struct {
	PathName string
	FileName string
}

type PathTransFormFunc func(string) PathKey

type StoreOpts struct {
	PathTransFormFunc PathTransFormFunc
	Root              string
}

type Store struct {
	StoreOpts
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		FileName: key,
	}
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

	if opts.PathTransFormFunc == nil {
		opts.PathTransFormFunc = DefaultPathTransformFunc
	}
	//or opts.Root == " "
	if len(opts.Root) == 0 {
		opts.Root = DefaultRootName
	}

	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Has(key string) bool {

	pathKey := CASPathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)

	_, err := os.Stat(pathNameWithRoot)
	if errors.Is(err, fs.ErrNotExist) {
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

func (s *Store) clearAll() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}

func (s *Store) WriteDecrypt(encKey []byte, key string, r io.Reader) (int64, error) {

	pathKey := CASPathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)

	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		fmt.Printf("An Error occured:%s ", err)
		return 0, err
	}

	//we make a PathKey struct, so we can easily access the hashStr without /
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, pathKey.PathName, pathKey.FileName)

	f, err := os.Create(fullPathWithRoot)
	if err != nil {
		return 0, err
	}

	//f is the returned file that is created
	n, err := encryption.CopyDecrypt(encKey, r, f)
	if err != nil {
		return 0, err
	}

	fmt.Printf("Written %d bytes to the disk %s\n", n, fullPathWithRoot)

	return n, nil

}

func (s *Store) writeStream(key string, r io.Reader) (int64, error) {

	pathKey := CASPathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)

	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		fmt.Printf("An Error occured:%s ", err)
		return 0, err
	}

	//we make a PathKey struct, so we can easily access the hashStr without /
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, pathKey.PathName, pathKey.FileName)

	f, err := os.Create(fullPathWithRoot)
	if err != nil {
		return 0, err
	}

	defer f.Close()
	//f is the returned file that is created
	n, err := io.Copy(f, r)
	if err != nil {
		return 0, err
	}

	fmt.Printf("Written %d bytes to the disk %s\n", n, fullPathWithRoot)

	return n, nil
}

func (s *Store) Read(key string) (int64, io.Reader, error) {

	f, err := s.readStream(key)
	if err != nil {
		return 0, nil, err
	}

	buf := new(bytes.Buffer)
	n, err := io.Copy(buf, f)

	if err := f.Close(); err != nil {
		return 0, nil, err
	}
	return n, buf, err

}

// to give user more authority on what to do with the read data,
// how and when to read or close it, nah.., actually making this private
func (s *Store) readStream(key string) (io.ReadCloser, error) {

	pathKey := CASPathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, pathKey.PathName, pathKey.FileName)

	return os.Open(pathNameWithRoot)
}

func (s *Store) Delete(key string) error {

	pathKey := CASPathTransformFunc(key)

	defer func() {
		fmt.Printf("deleted [%s] from disk\n", pathKey.FileName)
	}()

	fmt.Printf("The root: %s\n", s.Root)
	fmt.Printf("The full path: %s\n", pathKey.fullPath())
	fmt.Printf("The first path: %s\n", pathKey.firstPath(pathKey.fullPath()))

	firstPath := fmt.Sprintf("%s/%s", s.Root, pathKey.firstPath(pathKey.fullPath()))

	fmt.Printf("The path name with root: %s\n", firstPath)
	return os.RemoveAll(firstPath)

}
