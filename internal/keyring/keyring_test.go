package keyring

import (
	"errors"
	"testing"

	gokeyring "github.com/zalando/go-keyring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	gokeyring.MockInit()

	err := Set("test-service", "test-user", "test-secret")
	require.NoError(t, err)

	// Verify it was stored
	val, err := Get("test-service", "test-user")
	require.NoError(t, err)
	assert.Equal(t, "test-secret", val)
}

func TestGet(t *testing.T) {
	t.Run("returns stored secret", func(t *testing.T) {
		gokeyring.MockInit()

		// Store a secret first
		err := gokeyring.Set("test-service", "test-user", "my-secret")
		require.NoError(t, err)

		val, err := Get("test-service", "test-user")
		require.NoError(t, err)
		assert.Equal(t, "my-secret", val)
	})

	t.Run("returns ErrNotFound for missing secret", func(t *testing.T) {
		gokeyring.MockInit()

		_, err := Get("test-service", "nonexistent-user")
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestDelete(t *testing.T) {
	gokeyring.MockInit()

	// Store a secret
	err := Set("test-service", "test-user", "secret-to-delete")
	require.NoError(t, err)

	// Delete it
	err = Delete("test-service", "test-user")
	require.NoError(t, err)

	// Verify it's gone
	_, err = Get("test-service", "test-user")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestSetWithError(t *testing.T) {
	mockErr := errors.New("keyring unavailable")
	gokeyring.MockInitWithError(mockErr)

	err := Set("test-service", "test-user", "secret")
	assert.ErrorIs(t, err, mockErr)
}

func TestGetWithError(t *testing.T) {
	mockErr := errors.New("keyring unavailable")
	gokeyring.MockInitWithError(mockErr)

	_, err := Get("test-service", "test-user")
	assert.ErrorIs(t, err, mockErr)
}

func TestDeleteWithError(t *testing.T) {
	mockErr := errors.New("keyring unavailable")
	gokeyring.MockInitWithError(mockErr)

	err := Delete("test-service", "test-user")
	assert.ErrorIs(t, err, mockErr)
}

func TestTimeoutError(t *testing.T) {
	err := &TimeoutError{message: "test timeout"}
	assert.Equal(t, "test timeout", err.Error())
}
