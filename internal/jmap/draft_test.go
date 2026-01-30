package jmap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuoteText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single line",
			input:    "Hello world",
			expected: "> Hello world",
		},
		{
			name:     "multiple lines",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "> Line 1\n> Line 2\n> Line 3",
		},
		{
			name:     "already quoted text",
			input:    "> Previously quoted\nNew line",
			expected: "> > Previously quoted\n> New line",
		},
		{
			name:     "empty lines preserved",
			input:    "First\n\nThird",
			expected: "> First\n> \n> Third",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quoteText(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatReplyHTML(t *testing.T) {
	t.Run("with HTML original", func(t *testing.T) {
		result := formatReplyHTML(
			"Thanks!",
			"On Jan 30, 2026, alice@example.com wrote:",
			"<p>Original <b>HTML</b> content</p>",
			"Original text content",
		)

		// Should be a full HTML document
		assert.Contains(t, result, "<!DOCTYPE html>")
		assert.Contains(t, result, "<html>")
		assert.Contains(t, result, "</html>")
		// Should contain reply text in div
		assert.Contains(t, result, "<div>Thanks!</div>")
		// Should contain attribution
		assert.Contains(t, result, "On Jan 30, 2026, alice@example.com wrote:")
		// Should use HTML original in blockquote
		assert.Contains(t, result, "<blockquote type=\"cite\" id=\"qt\">")
		assert.Contains(t, result, "<p>Original <b>HTML</b> content</p>")
		// Should NOT use text version when HTML is available
		assert.NotContains(t, result, "Original text content")
		// Should have blank line before attribution
		assert.Contains(t, result, "</div><div><br></div><div>On Jan 30")
	})

	t.Run("with text only original", func(t *testing.T) {
		result := formatReplyHTML(
			"Thanks!",
			"On Jan 30, 2026, bob@example.com wrote:",
			"", // no HTML
			"Plain text original",
		)

		assert.Contains(t, result, "<div>Thanks!</div>")
		assert.Contains(t, result, "<blockquote type=\"cite\" id=\"qt\">")
		assert.Contains(t, result, "<div>Plain text original</div>")
	})

	t.Run("escapes HTML in reply text", func(t *testing.T) {
		result := formatReplyHTML(
			"Check <script>alert('xss')</script>",
			"On Jan 30, 2026, test@example.com wrote:",
			"",
			"Original",
		)

		// Should escape dangerous HTML
		assert.Contains(t, result, "&lt;script&gt;")
		assert.NotContains(t, result, "<script>alert")
	})

	t.Run("converts line breaks to divs", func(t *testing.T) {
		result := formatReplyHTML(
			"Line 1\nLine 2",
			"Attribution",
			"",
			"Original",
		)

		assert.Contains(t, result, "<div>Line 1</div><div>Line 2</div>")
	})

	t.Run("converts empty lines to br divs", func(t *testing.T) {
		result := formatReplyHTML(
			"Line 1\n\nLine 3",
			"Attribution",
			"",
			"Original",
		)

		assert.Contains(t, result, "<div>Line 1</div><div><br></div><div>Line 3</div>")
	})

	t.Run("escapes attribution", func(t *testing.T) {
		result := formatReplyHTML(
			"Reply",
			"On Jan 30, <attacker@evil.com> wrote:",
			"",
			"Original",
		)

		assert.Contains(t, result, "&lt;attacker@evil.com&gt;")
	})

	t.Run("has blank line after blockquote", func(t *testing.T) {
		result := formatReplyHTML(
			"Reply",
			"Attribution",
			"<p>Original</p>",
			"",
		)

		assert.Contains(t, result, "</blockquote><div><br></div></body>")
	})
}

func TestTextToHTMLDivs(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		assert.Equal(t, "", textToHTMLDivs(""))
	})

	t.Run("single line", func(t *testing.T) {
		assert.Equal(t, "<div>Hello</div>", textToHTMLDivs("Hello"))
	})

	t.Run("multiple lines", func(t *testing.T) {
		result := textToHTMLDivs("Line 1\nLine 2\nLine 3")
		assert.Equal(t, "<div>Line 1</div><div>Line 2</div><div>Line 3</div>", result)
	})

	t.Run("empty lines become br divs", func(t *testing.T) {
		result := textToHTMLDivs("First\n\nThird")
		assert.Equal(t, "<div>First</div><div><br></div><div>Third</div>", result)
	})

	t.Run("escapes HTML", func(t *testing.T) {
		result := textToHTMLDivs("<script>alert('xss')</script>")
		assert.Equal(t, "<div>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</div>", result)
	})
}

func TestCreateReplyDraftQuoting(t *testing.T) {
	// Test that the plain text body includes quoted original
	t.Run("plain text includes attribution and quoted text", func(t *testing.T) {
		// This is a unit test for the text formatting logic
		body := "My reply"
		attribution := "On Mon, Jan 30, 2026 at 3:04 PM, alice@example.com wrote:"
		originalText := "Original message\nSecond line"

		textBody := body + "\n\n" + attribution + "\n" + quoteText(originalText)

		assert.True(t, strings.HasPrefix(textBody, "My reply"))
		assert.Contains(t, textBody, attribution)
		assert.Contains(t, textBody, "> Original message")
		assert.Contains(t, textBody, "> Second line")
	})
}
