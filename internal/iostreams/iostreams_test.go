package iostreams

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTest(t *testing.T) {
	ios, stdin, stdout, stderr := Test()

	assert.NotNil(t, ios)
	assert.NotNil(t, stdin)
	assert.NotNil(t, stdout)
	assert.NotNil(t, stderr)

	// Test streams should not be TTYs
	assert.False(t, ios.IsStdinTTY())
	assert.False(t, ios.IsStdoutTTY())
	assert.False(t, ios.IsStderrTTY())
	assert.False(t, ios.IsInteractive())

	// Can write to stdin buffer and read from streams
	stdin.WriteString("test input")
	buf := make([]byte, 10)
	n, _ := ios.In.Read(buf)
	assert.Equal(t, "test input", string(buf[:n]))

	// Can write to stdout/stderr through ios
	ios.Out.Write([]byte("stdout"))
	ios.ErrOut.Write([]byte("stderr"))
	assert.Equal(t, "stdout", stdout.String())
	assert.Equal(t, "stderr", stderr.String())
}

func TestIsInteractive(t *testing.T) {
	t.Run("false when stdin not TTY", func(t *testing.T) {
		ios := &IOStreams{
			stdinIsTTY:  false,
			stdoutIsTTY: true,
		}
		assert.False(t, ios.IsInteractive())
	})

	t.Run("false when stdout not TTY", func(t *testing.T) {
		ios := &IOStreams{
			stdinIsTTY:  true,
			stdoutIsTTY: false,
		}
		assert.False(t, ios.IsInteractive())
	})

	t.Run("true when both TTY", func(t *testing.T) {
		ios := &IOStreams{
			stdinIsTTY:  true,
			stdoutIsTTY: true,
		}
		assert.True(t, ios.IsInteractive())
	})
}

func TestIsSafeMode(t *testing.T) {
	t.Run("true when stdin not TTY", func(t *testing.T) {
		ios := &IOStreams{
			stdinIsTTY: false,
		}
		assert.True(t, ios.IsSafeMode())
	})

	t.Run("false when stdin is TTY", func(t *testing.T) {
		ios := &IOStreams{
			stdinIsTTY: true,
		}
		assert.False(t, ios.IsSafeMode())
	})

	t.Run("false when FM_UNSAFE=1", func(t *testing.T) {
		t.Setenv("FM_UNSAFE", "1")

		ios := &IOStreams{
			stdinIsTTY: false, // Would normally be safe mode
		}
		assert.False(t, ios.IsSafeMode())
	})

	t.Run("still safe when FM_UNSAFE is not 1", func(t *testing.T) {
		t.Setenv("FM_UNSAFE", "true") // Not "1"

		ios := &IOStreams{
			stdinIsTTY: false,
		}
		assert.True(t, ios.IsSafeMode())
	})
}

func TestColorEnabled(t *testing.T) {
	t.Run("false when stdout not TTY", func(t *testing.T) {
		ios := &IOStreams{
			stdoutIsTTY: false,
		}
		assert.False(t, ios.ColorEnabled())
	})

	t.Run("true when stdout is TTY", func(t *testing.T) {
		ios := &IOStreams{
			stdoutIsTTY: true,
		}
		assert.True(t, ios.ColorEnabled())
	})

	t.Run("false when NO_COLOR set", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")

		ios := &IOStreams{
			stdoutIsTTY: true,
		}
		assert.False(t, ios.ColorEnabled())
	})

	t.Run("false when FM_NO_COLOR set", func(t *testing.T) {
		t.Setenv("FM_NO_COLOR", "1")

		ios := &IOStreams{
			stdoutIsTTY: true,
		}
		assert.False(t, ios.ColorEnabled())
	})

	t.Run("caches result", func(t *testing.T) {
		ios := &IOStreams{
			stdoutIsTTY: true,
		}

		// First call caches the result
		result1 := ios.ColorEnabled()
		assert.True(t, result1)
		assert.True(t, ios.colorChecked)

		// Change the TTY flag - should not affect cached result
		ios.stdoutIsTTY = false
		result2 := ios.ColorEnabled()
		assert.True(t, result2) // Still true from cache
	})
}

func TestSetColorEnabled(t *testing.T) {
	ios := &IOStreams{
		stdoutIsTTY: false,
	}

	// Without SetColorEnabled, would be false
	assert.False(t, ios.ColorEnabled())

	// Reset and force enable
	ios.colorChecked = false
	ios.SetColorEnabled(true)

	assert.True(t, ios.ColorEnabled())
	assert.True(t, ios.colorChecked)
}

func TestTerminalWidth(t *testing.T) {
	t.Run("returns 80 when not TTY", func(t *testing.T) {
		ios := &IOStreams{
			stdoutIsTTY: false,
		}
		assert.Equal(t, 80, ios.TerminalWidth())
	})
}
