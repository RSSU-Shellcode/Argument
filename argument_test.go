package argument

import (
	"bytes"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/For-ACGN/monkey"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		arg0 := []byte{0x12, 0x34, 0x56, 0x78}
		arg1 := bytes.Repeat([]byte("hello runtime"), 10)
		arg2 := make([]byte, 0)
		args := [][]byte{arg0, arg1, arg2}
		stub, err := Encode(args)
		require.NoError(t, err)

		header := 44
		argSize := 3 * 4
		argLen := len(arg0) + len(arg1)
		expected := header + argSize + argLen
		require.Len(t, stub, expected)
	})

	t.Run("failed to generate crypto key", func(t *testing.T) {
		patch := func(b []byte) (int, error) {
			return 0, errors.New("monkey error")
		}
		pg := monkey.Patch(rand.Read, patch)
		defer pg.Unpatch()

		arg0 := []byte{0x12, 0x34, 0x56, 0x78}
		stub, err := Encode([][]byte{arg0})
		require.Error(t, err)
		require.Nil(t, stub)
	})
}

func TestDecode(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		arg0 := []byte{0x12, 0x34, 0x56, 0x78}
		arg1 := bytes.Repeat([]byte("hello runtime"), 10)
		arg2 := make([]byte, 0)
		args := [][]byte{arg0, arg1, arg2}
		stub, err := Encode(args)
		require.NoError(t, err)

		output, err := Decode(stub)
		require.NoError(t, err)
		require.Equal(t, args, output)
	})

	t.Run("short stub", func(t *testing.T) {
		stub, err := Decode(nil)
		require.EqualError(t, err, "stub is too short")
		require.Nil(t, stub)
	})

	t.Run("no argument", func(t *testing.T) {
		stub, err := Encode(nil)
		require.NoError(t, err)

		output, err := Decode(stub)
		require.NoError(t, err)
		require.Empty(t, output)
	})

	t.Run("invalid checksum", func(t *testing.T) {
		arg0 := []byte{0x12, 0x34, 0x56, 0x78}
		arg1 := bytes.Repeat([]byte("hello runtime"), 10)
		arg2 := make([]byte, 0)
		args := [][]byte{arg0, arg1, arg2}
		stub, err := Encode(args)
		require.NoError(t, err)

		// destruct checksum
		copy(stub[offsetChecksum:], []byte{0x00, 0x00, 0x00, 0x00})

		output, err := Decode(stub)
		require.EqualError(t, err, "invalid checksum")
		require.Nil(t, output)
	})
}
