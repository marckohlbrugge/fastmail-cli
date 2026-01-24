package iostreams

import (
	"io"
	"os"

	"golang.org/x/term"
)

// IOStreams provides access to standard input/output streams.
type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer

	stdinIsTTY  bool
	stdoutIsTTY bool
	stderrIsTTY bool

	colorEnabled bool
	colorChecked bool
}

// System returns IOStreams configured for the standard system streams.
func System() *IOStreams {
	stdinFd := int(os.Stdin.Fd())
	stdoutFd := int(os.Stdout.Fd())
	stderrFd := int(os.Stderr.Fd())

	return &IOStreams{
		In:          os.Stdin,
		Out:         os.Stdout,
		ErrOut:      os.Stderr,
		stdinIsTTY:  term.IsTerminal(stdinFd),
		stdoutIsTTY: term.IsTerminal(stdoutFd),
		stderrIsTTY: term.IsTerminal(stderrFd),
	}
}

// Test returns IOStreams for testing with the provided readers/writers.
func Test() *IOStreams {
	return &IOStreams{
		In:          &nullReader{},
		Out:         io.Discard,
		ErrOut:      io.Discard,
		stdinIsTTY:  false,
		stdoutIsTTY: false,
		stderrIsTTY: false,
	}
}

// IsStdinTTY returns true if stdin is connected to a terminal.
func (s *IOStreams) IsStdinTTY() bool {
	return s.stdinIsTTY
}

// IsStdoutTTY returns true if stdout is connected to a terminal.
func (s *IOStreams) IsStdoutTTY() bool {
	return s.stdoutIsTTY
}

// IsStderrTTY returns true if stderr is connected to a terminal.
func (s *IOStreams) IsStderrTTY() bool {
	return s.stderrIsTTY
}

// IsInteractive returns true if both stdin and stdout are connected to terminals.
func (s *IOStreams) IsInteractive() bool {
	return s.stdinIsTTY && s.stdoutIsTTY
}

// IsSafeMode returns true when destructive operations should be blocked.
// Safe mode is active when stdin is not a terminal (AI/script execution).
// Can be overridden via FM_UNSAFE=1 environment variable.
func (s *IOStreams) IsSafeMode() bool {
	if os.Getenv("FM_UNSAFE") == "1" {
		return false
	}
	return !s.stdinIsTTY
}

// ColorEnabled returns true if color output is enabled.
// Respects NO_COLOR and FM_NO_COLOR environment variables.
func (s *IOStreams) ColorEnabled() bool {
	if !s.colorChecked {
		s.colorChecked = true
		s.colorEnabled = s.stdoutIsTTY &&
			os.Getenv("NO_COLOR") == "" &&
			os.Getenv("FM_NO_COLOR") == ""
	}
	return s.colorEnabled
}

// SetColorEnabled forces color output on or off.
func (s *IOStreams) SetColorEnabled(enabled bool) {
	s.colorChecked = true
	s.colorEnabled = enabled
}

// TerminalWidth returns the width of the terminal, or 80 if not a TTY.
func (s *IOStreams) TerminalWidth() int {
	if !s.stdoutIsTTY {
		return 80
	}

	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

type nullReader struct{}

func (r *nullReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}
