package cmdutil

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/marckohlbrugge/fastmail-cli/internal/jmap"
)

// DefaultEmailFields are the fields shown in human output.
var DefaultEmailFields = []string{"id", "date", "from", "subject"}

// AvailableEmailFields lists all fields that can be displayed.
var AvailableEmailFields = []string{"id", "threadId", "subject", "from", "to", "cc", "date", "preview", "unread", "attachment"}

// FieldConfig defines display width for a field.
type FieldConfig struct {
	Width  int
	Getter func(jmap.Email) string
}

// EmailFieldConfigs maps field names to their display configuration.
var EmailFieldConfigs = map[string]FieldConfig{
	"id":         {Width: 12, Getter: func(e jmap.Email) string { return e.ID }},
	"threadId":   {Width: 12, Getter: func(e jmap.Email) string { return e.ThreadID }},
	"subject":    {Width: 50, Getter: func(e jmap.Email) string { return e.Subject }},
	"from":       {Width: 30, Getter: func(e jmap.Email) string { return formatAddresses(e.From) }},
	"to":         {Width: 30, Getter: func(e jmap.Email) string { return formatAddresses(e.To) }},
	"cc":         {Width: 30, Getter: func(e jmap.Email) string { return formatAddresses(e.CC) }},
	"date":       {Width: 12, Getter: func(e jmap.Email) string { return FormatRelativeDate(e.ReceivedAt) }},
	"preview":    {Width: 60, Getter: func(e jmap.Email) string { return e.Preview }},
	"unread":     {Width: 1, Getter: func(e jmap.Email) string { if e.IsUnread() { return "*" }; return " " }},
	"attachment": {Width: 1, Getter: func(e jmap.Email) string { if e.HasAttachment { return "+" }; return " " }},
}

// ParseFields parses a comma-separated fields string, returning defaults if empty.
func ParseFields(fieldsArg string) []string {
	if fieldsArg == "" {
		return DefaultEmailFields
	}
	fields := strings.Split(fieldsArg, ",")
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}
	return fields
}

// ValidateFields checks that all requested fields are valid.
func ValidateFields(fields []string) error {
	for _, f := range fields {
		if _, ok := EmailFieldConfigs[f]; !ok {
			return fmt.Errorf("unknown field %q, available: %s", f, strings.Join(AvailableEmailFields, ", "))
		}
	}
	return nil
}

// FormatEmailRow formats a single email according to the specified fields.
func FormatEmailRow(email jmap.Email, fields []string) string {
	var parts []string
	for _, field := range fields {
		config := EmailFieldConfigs[field]
		value := config.Getter(email)
		if value == "" {
			if field == "subject" {
				value = "(no subject)"
			} else if field == "from" || field == "to" || field == "cc" {
				value = "(unknown)"
			}
		}
		value = Truncate(value, config.Width)
		parts = append(parts, fmt.Sprintf("%-*s", config.Width, value))
	}
	return strings.Join(parts, "  ")
}

// PrintEmailList prints a list of emails with the specified fields.
func PrintEmailList(out io.Writer, emails []jmap.Email, fields []string) {
	for _, email := range emails {
		fmt.Fprintln(out, FormatEmailRow(email, fields))
	}
}

// Truncate shortens a string to maxLen, adding ellipsis if needed.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

// FormatRelativeDate formats a time as a relative date string.
func FormatRelativeDate(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins <= 1 {
			return "just now"
		}
		return fmt.Sprintf("%dm ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	case diff < 48*time.Hour:
		return "Yesterday"
	case diff < 7*24*time.Hour:
		return t.Weekday().String()[:3]
	case t.Year() == now.Year():
		return t.Format("Jan 2")
	default:
		return t.Format("Jan 2, 2006")
	}
}

func formatAddresses(addrs []jmap.EmailAddress) string {
	if len(addrs) == 0 {
		return ""
	}
	var names []string
	for _, a := range addrs {
		if a.Name != "" {
			names = append(names, a.Name)
		} else {
			parts := strings.Split(a.Email, "@")
			names = append(names, parts[0])
		}
	}
	return strings.Join(names, ", ")
}
