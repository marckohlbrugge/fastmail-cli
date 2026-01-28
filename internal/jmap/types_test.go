package jmap

import "testing"

func TestEmail_IsUnread(t *testing.T) {
	tests := []struct {
		name     string
		keywords map[string]bool
		want     bool
	}{
		{
			name:     "nil keywords is unread",
			keywords: nil,
			want:     true,
		},
		{
			name:     "empty keywords is unread",
			keywords: map[string]bool{},
			want:     true,
		},
		{
			name:     "seen keyword means read",
			keywords: map[string]bool{"$seen": true},
			want:     false,
		},
		{
			name:     "other keywords without seen is unread",
			keywords: map[string]bool{"$flagged": true, "$draft": true},
			want:     true,
		},
		{
			name:     "seen false is unread",
			keywords: map[string]bool{"$seen": false},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Email{Keywords: tt.keywords}
			if got := e.IsUnread(); got != tt.want {
				t.Errorf("IsUnread() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmail_IsDraft(t *testing.T) {
	tests := []struct {
		name     string
		keywords map[string]bool
		want     bool
	}{
		{
			name:     "nil keywords is not draft",
			keywords: nil,
			want:     false,
		},
		{
			name:     "empty keywords is not draft",
			keywords: map[string]bool{},
			want:     false,
		},
		{
			name:     "draft keyword means draft",
			keywords: map[string]bool{"$draft": true},
			want:     true,
		},
		{
			name:     "draft false is not draft",
			keywords: map[string]bool{"$draft": false},
			want:     false,
		},
		{
			name:     "other keywords without draft is not draft",
			keywords: map[string]bool{"$seen": true, "$flagged": true},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Email{Keywords: tt.keywords}
			if got := e.IsDraft(); got != tt.want {
				t.Errorf("IsDraft() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailAddress_String(t *testing.T) {
	tests := []struct {
		name  string
		addr  EmailAddress
		want  string
	}{
		{
			name:  "email only",
			addr:  EmailAddress{Email: "test@example.com"},
			want:  "test@example.com",
		},
		{
			name:  "name and email",
			addr:  EmailAddress{Name: "John Doe", Email: "john@example.com"},
			want:  "John Doe <john@example.com>",
		},
		{
			name:  "empty name treated as email only",
			addr:  EmailAddress{Name: "", Email: "test@example.com"},
			want:  "test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.addr.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatAddresses(t *testing.T) {
	tests := []struct {
		name  string
		addrs []EmailAddress
		want  string
	}{
		{
			name:  "empty slice",
			addrs: []EmailAddress{},
			want:  "",
		},
		{
			name:  "nil slice",
			addrs: nil,
			want:  "",
		},
		{
			name: "single address",
			addrs: []EmailAddress{
				{Name: "Alice", Email: "alice@example.com"},
			},
			want: "Alice <alice@example.com>",
		},
		{
			name: "multiple addresses",
			addrs: []EmailAddress{
				{Name: "Alice", Email: "alice@example.com"},
				{Email: "bob@example.com"},
			},
			want: "Alice <alice@example.com>, bob@example.com",
		},
		{
			name: "three addresses",
			addrs: []EmailAddress{
				{Name: "Alice", Email: "alice@example.com"},
				{Name: "Bob", Email: "bob@example.com"},
				{Email: "charlie@example.com"},
			},
			want: "Alice <alice@example.com>, Bob <bob@example.com>, charlie@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatAddresses(tt.addrs); got != tt.want {
				t.Errorf("FormatAddresses() = %q, want %q", got, tt.want)
			}
		})
	}
}
