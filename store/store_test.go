package store

import (
	"bytes"
	"fmt"
	"runtime"

	"testing"
)

func TestStore(t *testing.T) {

	opts := StoreOpts{
		PathTransFormFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)

	for i := 0; i < 50; i++ {

		key := fmt.Sprintf("vinland_%d", i)
		data := []byte("Some jpeg file")

		_, err := s.Write(key, bytes.NewReader(data), "test")
		if err != nil {
			t.Error(err)
		}

		if ok := s.Has(key, "test"); !ok {
			t.Error("Expected to have the key\n")
		}

		_, _, err = s.Read(key, "test")
		if err != nil {
			t.Error(err)
		}

		// Force garbage collection to release file handles (Windows-specific)
		runtime.GC()

		// b, _ := ioutil.ReadAll(r)
		// if string(b) == string(data) {
		// 	t.Errorf("want %s have %s", data, b)
		// }

		if err := s.Delete(key, "test"); err != nil {
			t.Error(err)
		}

		// if ok := s.Has(key, "test"); ok {
		// 	t.Error("Expected to not have the key\n")
		// }
	}
	// pathKey := CASPathTransformFunc("vinland")

}

func TestDelete(t *testing.T) {
	opts := StoreOpts{
		PathTransFormFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)

	err := s.Delete("vinland", "test")
	if err != nil {
		t.Error(err)
	}
}

// func teardown(t *testing.T) {
// 	opts := StoreOpts{
// 		pathTransFormFunc: CASPathTransformFunc,
// 	}
// 	s := NewStore(opts)

// 	if err := s.clearAll(); err != nil {
// 		t.Error(err)
// 	}
// }
