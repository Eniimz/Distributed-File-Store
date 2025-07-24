package store

import (
	"bytes"

	"testing"
)

func TestStore(t *testing.T) {

	opts := StoreOpts{
		pathTransFormFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)

	// pathKey := CASPathTransformFunc("vinland")
	data := []byte("Some jpeg file")

	s.writeStream("vinland", bytes.NewReader(data))

}

func TestDelete(t *testing.T) {
	opts := StoreOpts{
		pathTransFormFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)

	err := s.Delete("vinland")
	if err != nil {
		t.Error(err)
	}
}

func TestTransformFunc(t *testing.T) {

}
