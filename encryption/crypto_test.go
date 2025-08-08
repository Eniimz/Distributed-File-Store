package encryption

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCopyEncrypt(t *testing.T) {
	src := bytes.NewReader([]byte("The big data file"))
	dst := new(bytes.Buffer)
	key := NewEncryptionKey()

	_, err := CopyEncrypt(key, src, dst)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Encrypted Output: %s\n", dst.String())

	out := new(bytes.Buffer)
	_, err = CopyDecrypt(key, dst, out)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Decrypted Output: %s\n", out.String())
}
