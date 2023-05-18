package log

import (
	"os"
	"testing"
)

func TestStoreAppendRead(t *testing.T) {
	f, err := os.CreateTemp("", "store_append_read_test")
}
