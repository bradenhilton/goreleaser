package logext

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/caarlos0/log"
	"github.com/goreleaser/goreleaser/v2/internal/golden"
	"github.com/stretchr/testify/require"
)

func TestWriter(t *testing.T) {
	t.Setenv("CI", "")

	t.Run("info", func(t *testing.T) {
		t.Cleanup(func() {
			log.Log = log.New(os.Stderr)
		})
		var b bytes.Buffer
		log.Log = log.New(&b)
		l, err := io.WriteString(NewWriter(), "foo\nbar\n")
		require.NoError(t, err)
		require.Equal(t, 8, l)
		require.Empty(t, b.String())
	})

	t.Run("debug", func(t *testing.T) {
		t.Cleanup(func() {
			log.Log = log.New(os.Stderr)
		})
		var b bytes.Buffer
		log.Log = log.New(&b)
		log.SetLevel(log.DebugLevel)
		l, err := io.WriteString(NewWriter(), "foo\nbar\n")
		require.NoError(t, err)
		require.Equal(t, 8, l)
		golden.RequireEqualTxt(t, b.Bytes())
	})
}
