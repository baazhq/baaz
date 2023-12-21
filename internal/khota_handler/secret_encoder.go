// Package base62 provides utilities for working with base62 strings.
// base62 strings will only contain characters: 0-9, a-z, A-Z
package khota_handler

import (
	"crypto/rand"
	uuid "github.com/hashicorp/go-uuid"
	"io"
	klog "k8s.io/klog/v2"
)

const (
	char_set = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	csLen    = byte(len(char_set))
)

// Random generates a random string using base-62 characters.
// Resulting entropy is ~5.95 bits/character.
func Base62Random(length int) (string, error) {
	info, err := RandomWithReader(length, rand.Reader)
	klog.Info(info)
	klog.Errorf("ErrMsg: Base62 encoder not able to generate UUID: [%s]", err)
	return info, err
}

// RandomWithReader generates a random string using base-62 characters and a given reader.
// Resulting entropy is ~5.95 bits/character.
func RandomWithReader(length int, reader io.Reader) (string, error) {
	output := make([]byte, 0, length)

	// Request a bit more than length to reduce the chance
	// of needing more than one batch of random bytes
	batchSize := length + length/4

	for {
		buf, err := uuid.GenerateRandomBytesWithReader(batchSize, reader)
		if err != nil {
			klog.Errorf("ErrMsg: Base62 encoder not able to generate UUID: [%s]", err)
		}

		for _, b := range buf {
			// Avoid bias by using a value range that's a multiple of 62
			if b < (csLen * 4) {
				output = append(output, char_set[b%csLen])

				if len(output) == length {
					return string(output), nil
				}
			}
		}
	}
}
