package log

import (
	"os"

	"github.com/stretchr/testify/require"
)

func TestIndex(t *testing.T){
	f, err := os.CreateTemp(os.TempDir(), "index_test")
	require.NoError(t, err)
	defer os.Remove()
}