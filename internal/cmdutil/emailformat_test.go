package cmdutil

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
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
			got := ParseFields(tt.fieldsArg)
			if len(got) != len(tt.want) {
				t.Errorf("ParseFields() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ParseFields()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFields() error = %v, wantErr %v", err, tt.wantErr)
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
			if got := Truncate(tt.s, tt.maxLen); got != tt.want {
				t.Errorf("Truncate() = %q, want %q", got, tt.want)
			}
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
			got := FormatRelativeDate(tt.t)
			if got != tt.want {
				t.Errorf("FormatRelativeDate() = %q, want %q", got, tt.want)
			}
		})
	}

	// Test previous year separately to avoid date calculation issues
	t.Run("previous year", func(t *testing.T) {
		pastYear := time.Date(now.Year()-1, 6, 15, 10, 0, 0, 0, time.Local)
		got := FormatRelativeDate(pastYear)
		want := "Jun 15, " + pastYear.Format("2006")
		if got != want {
			t.Errorf("FormatRelativeDate() = %q, want %q", got, want)
		}
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
			if got := formatAddresses(tt.addrs); got != tt.want {
				t.Errorf("formatAddresses() = %q, want %q", got, tt.want)
			}
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
		if !strings.Contains(row, "abc123") {
			t.Error("row should contain email ID")
		}
		if !strings.Contains(row, "Test Subject") {
			t.Error("row should contain subject")
		}
		if !strings.Contains(row, "Sender") {
			t.Error("row should contain sender name")
		}
	})

	t.Run("formats unread indicator", func(t *testing.T) {
		row := FormatEmailRow(email, []string{"unread"})
		if !strings.Contains(row, "*") {
			t.Error("unread email should show *")
		}
	})

	t.Run("formats no subject placeholder", func(t *testing.T) {
		noSubject := jmap.Email{ID: "123", Subject: ""}
		row := FormatEmailRow(noSubject, []string{"subject"})
		if !strings.Contains(row, "(no subject)") {
			t.Error("empty subject should show placeholder")
		}
	})

	t.Run("formats unknown sender placeholder", func(t *testing.T) {
		noFrom := jmap.Email{ID: "123", From: nil}
		row := FormatEmailRow(noFrom, []string{"from"})
		if !strings.Contains(row, "(unknown)") {
			t.Error("empty from should show placeholder")
		}
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

	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "email1") {
		t.Error("first line should contain email1")
	}
	if !strings.Contains(lines[1], "email2") {
		t.Error("second line should contain email2")
	}
}
