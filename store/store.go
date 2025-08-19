package store

import (
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

func (s *Store) Has(key string, id string) bool {

	pathKey := CASPathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.PathName)

	fmt.Printf("Checking if file exists: %s\n", pathNameWithRoot)
	_, err := os.Stat(pathNameWithRoot)
	return !errors.Is(err, fs.ErrNotExist)
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

func (s *Store) Write(key string, r io.Reader, id string) (int64, error) {
	return s.writeStream(key, r, id)
}

func (s *Store) WriteDecrypt(encKey []byte, key string, r io.Reader, id string) (int64, error) {

	f, err := s.openForWriting(key, id)
	if err != nil {
		return 0, err
	}

	defer f.Close()
	//f is the returned file that is created
	n, err := encryption.CopyDecrypt(encKey, r, f)
	if err != nil {
		return 0, err
	}

	fmt.Printf("Written %d bytes to the disk %s\n", n, f.Name())

	return n, nil

}

func (s *Store) openForWriting(key string, id string) (*os.File, error) {

	pathKey := CASPathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.PathName)

	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		fmt.Printf("An Error occured:%s ", err)
		return nil, err
	}

	//we make a PathKey struct, so we can easily access the hashStr without /
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s/%s", s.Root, id, pathKey.PathName, pathKey.FileName)

	return os.Create(fullPathWithRoot)

}

func (s *Store) writeStream(key string, r io.Reader, id string) (int64, error) {

	f, err := s.openForWriting(key, id)
	if err != nil {
		return 0, err
	}

	defer f.Close()

	fmt.Printf("Copying to file data to file: %s\n", f.Name())
	//f is the returned file that is created
	n, err := io.Copy(f, r)
	if err != nil {
		fmt.Printf("Error copying to file: %s\n", err)
		return 0, err
	}

	fmt.Printf("\n\nWritten %d bytes to the disk %s\n", n, f.Name())

	return n, nil
}

func (s *Store) Read(key string, id string) (int64, io.Reader, error) {

	return s.readStream(key, id)
}

// to give user more authority on what to do with the read data,
// how and when to read or close it, nah.., actually making this private
func (s *Store) readStream(key string, id string) (int64, io.ReadCloser, error) {

	pathKey := CASPathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s/%s", s.Root, id, pathKey.PathName, pathKey.FileName)

	file, err := os.Open(pathNameWithRoot)
	if err != nil {
		return 0, nil, err
	}

	fi, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}

	return fi.Size(), file, nil
}

func (s *Store) Delete(key string, id string) error {

	pathKey := CASPathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s/%s", s.Root, id, pathKey.PathName, pathKey.FileName)

	if !s.Has(key, id) {
		fmt.Printf("the file doesnt exist in the disk: %s", pathNameWithRoot)
		return nil
	}

	fmt.Printf("The file does exists in: %s", pathNameWithRoot)

	fmt.Printf("Deleting....")
	defer func() {
		fmt.Printf("deleted [%s] from disk\n", pathKey.FileName)
	}()

	firstPath := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.firstPath(pathKey.fullPath()))

	return os.RemoveAll(firstPath)

}
