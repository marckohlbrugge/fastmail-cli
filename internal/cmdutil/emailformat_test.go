package cmdutil

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFields(t *testing.T) {
	tests := []struct {
		name      string
		fieldsArg string
		want      []string
	}{
		{
			name:      "empty returns defaults",
			fieldsArg: "",
			want:      DefaultEmailFields,
		},
		{
			name:      "single field",
			fieldsArg: "subject",
			want:      []string{"subject"},
		},
		{
			name:      "multiple fields",
			fieldsArg: "id,subject,from",
			want:      []string{"id", "subject", "from"},
		},
		{
			name:      "fields with spaces",
			fieldsArg: "id, subject , from",
			want:      []string{"id", "subject", "from"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ParseFields(tt.fieldsArg))
		})
	}
}

func TestValidateFields(t *testing.T) {
	tests := []struct {
		name    string
		fields  []string
		wantErr bool
	}{
		{
			name:    "valid single field",
			fields:  []string{"id"},
			wantErr: false,
		},
		{
			name:    "valid multiple fields",
			fields:  []string{"id", "subject", "from", "date"},
			wantErr: false,
		},
		{
			name:    "all available fields",
			fields:  AvailableEmailFields,
			wantErr: false,
		},
		{
			name:    "invalid field",
			fields:  []string{"notafield"},
			wantErr: true,
		},
		{
			name:    "mix of valid and invalid",
			fields:  []string{"id", "invalid", "subject"},
			wantErr: true,
		},
		{
			name:    "empty slice is valid",
			fields:  []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFields(tt.fields)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		maxLen int
		want   string
	}{
		{
			name:   "string shorter than max",
			s:      "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "string equal to max",
			s:      "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "string longer than max",
			s:      "hello world",
			maxLen: 8,
			want:   "hello w…",
		},
		{
			name:   "empty string",
			s:      "",
			maxLen: 5,
			want:   "",
		},
		{
			name:   "max length 1",
			s:      "hello",
			maxLen: 1,
			want:   "…",
		},
		{
			name:   "unicode string (byte-based truncation)",
			s:      "héllo wörld",
			maxLen: 8,
			want:   "héllo …", // Note: uses byte length, not rune count
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Truncate(tt.s, tt.maxLen))
		})
	}
}

func TestFormatRelativeDate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{
			name: "just now (30 seconds ago)",
			t:    now.Add(-30 * time.Second),
			want: "just now",
		},
		{
			name: "just now (1 minute ago)",
			t:    now.Add(-1 * time.Minute),
			want: "just now",
		},
		{
			name: "minutes ago",
			t:    now.Add(-15 * time.Minute),
			want: "15m ago",
		},
		{
			name: "hours ago",
			t:    now.Add(-3 * time.Hour),
			want: "3h ago",
		},
		{
			name: "yesterday",
			t:    now.Add(-30 * time.Hour),
			want: "Yesterday",
		},
		{
			name: "weekday (4 days ago)",
			t:    now.Add(-4 * 24 * time.Hour),
			want: now.Add(-4 * 24 * time.Hour).Weekday().String()[:3],
		},
		{
			name: "same year older than a week",
			t:    time.Date(now.Year(), 1, 15, 10, 0, 0, 0, time.Local),
			want: "Jan 15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatRelativeDate(tt.t))
		})
	}

	// Test previous year separately to avoid date calculation issues
	t.Run("previous year", func(t *testing.T) {
		pastYear := time.Date(now.Year()-1, 6, 15, 10, 0, 0, 0, time.Local)
		want := "Jun 15, " + pastYear.Format("2006")
		assert.Equal(t, want, FormatRelativeDate(pastYear))
	})
}

func TestFormatAddresses(t *testing.T) {
	tests := []struct {
		name  string
		addrs []jmap.EmailAddress
		want  string
	}{
		{
			name:  "empty slice",
			addrs: []jmap.EmailAddress{},
			want:  "",
		},
		{
			name:  "nil slice",
			addrs: nil,
			want:  "",
		},
		{
			name: "single address with name",
			addrs: []jmap.EmailAddress{
				{Name: "Alice Smith", Email: "alice@example.com"},
			},
			want: "Alice Smith",
		},
		{
			name: "single address without name",
			addrs: []jmap.EmailAddress{
				{Email: "bob@example.com"},
			},
			want: "bob",
		},
		{
			name: "multiple addresses mixed",
			addrs: []jmap.EmailAddress{
				{Name: "Alice", Email: "alice@example.com"},
				{Email: "bob@example.com"},
			},
			want: "Alice, bob",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, formatAddresses(tt.addrs))
		})
	}
}

func TestFormatEmailRow(t *testing.T) {
	email := jmap.Email{
		ID:         "abc123",
		ThreadID:   "thread456",
		Subject:    "Test Subject",
		From:       []jmap.EmailAddress{{Name: "Sender", Email: "sender@example.com"}},
		ReceivedAt: time.Now().Add(-2 * time.Hour),
		Keywords:   map[string]bool{"$seen": false},
	}

	t.Run("formats with default fields", func(t *testing.T) {
		row := FormatEmailRow(email, DefaultEmailFields)
		assert.Contains(t, row, "abc123")
		assert.Contains(t, row, "Test Subject")
		assert.Contains(t, row, "Sender")
	})

	t.Run("formats unread indicator", func(t *testing.T) {
		row := FormatEmailRow(email, []string{"unread"})
		assert.Contains(t, row, "*")
	})

	t.Run("formats no subject placeholder", func(t *testing.T) {
		noSubject := jmap.Email{ID: "123", Subject: ""}
		row := FormatEmailRow(noSubject, []string{"subject"})
		assert.Contains(t, row, "(no subject)")
	})

	t.Run("formats unknown sender placeholder", func(t *testing.T) {
		noFrom := jmap.Email{ID: "123", From: nil}
		row := FormatEmailRow(noFrom, []string{"from"})
		assert.Contains(t, row, "(unknown)")
	})
}

func TestPrintEmailList(t *testing.T) {
	emails := []jmap.Email{
		{ID: "email1", Subject: "First"},
		{ID: "email2", Subject: "Second"},
	}

	var buf bytes.Buffer
	PrintEmailList(&buf, emails, []string{"id", "subject"})

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	require.Len(t, lines, 2)
	assert.Contains(t, lines[0], "email1")
	assert.Contains(t, lines[1], "email2")
}
