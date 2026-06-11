package htmlx

import "testing"

func TestHTMLToText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "plain text, no tags",
			in:   "Just plain text",
			want: "Just plain text",
		},
		{
			name: "empty string",
			in:   "",
			want: "",
		},
		{
			name: "adjacent paragraphs get a line break",
			in:   "<p>Backend</p><p>Engineer</p>",
			want: "Backend\nEngineer",
		},
		{
			name: "script subtree is dropped",
			in:   "<p>Hi</p><script>alert('x')</script><p>Bye</p>",
			want: "Hi\nBye",
		},
		{
			name: "style subtree is dropped",
			in:   "<style>.foo{color:red}</style><p>Visible</p>",
			want: "Visible",
		},
		{
			name: "inline tags do not break text",
			in:   "<p>Hello <strong>world</strong></p>",
			want: "Hello world",
		},
		{
			name: "nested blocks flatten without double blanks",
			in:   "<div><p>Hello</p></div>",
			want: "Hello",
		},
		{
			name: "br produces a line break",
			in:   "Line1<br>Line2",
			want: "Line1\nLine2",
		},
		{
			name: "list items each on their own line",
			in:   "<ul><li>One</li><li>Two</li></ul>",
			want: "One\nTwo",
		},
		{
			name: "headings break from following text",
			in:   "<h1>Title</h1><p>Body</p>",
			want: "Title\nBody",
		},
		{
			name: "html entities are decoded",
			in:   "<p>Tom &amp; Jerry &lt;dev&gt;</p>",
			want: "Tom & Jerry <dev>",
		},
		{
			name: "collapses runs of whitespace",
			in:   "<p>  Multiple   spaces  </p>",
			want: "Multiple spaces",
		},
		{
			name: "source newlines and indentation are normalized",
			in:   "<p>\n\t\tIndented line\n</p>",
			want: "Indented line",
		},
		{
			name: "realistic job description fragment",
			in:   `<div><h2>Senior Go Engineer</h2><p>We need someone with:</p><ul><li>Go &amp; Postgres</li><li>Docker</li></ul><script>track()</script></div>`,
			want: "Senior Go Engineer\nWe need someone with:\nGo & Postgres\nDocker",
		},
		{
			name: "table cells each on their own line",
			in:   "<table><tr><td>Go</td><td>Rust</td></tr></table>",
			want: "Go\nRust",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HTMLToText(tt.in)
			if err != nil {
				t.Fatalf("HTMLToText(%q) returned error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("HTMLToText(%q)\n  got:  %q\n  want: %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCollapseWhitespace(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"only whitespace", "   \n\t  \n  ", ""},
		{"trims and collapses", "  a   b  ", "a b"},
		{"drops blank lines", "a\n\n\nb", "a\nb"},
		{"preserves single breaks", "line1\nline2", "line1\nline2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collapseWhitespace(tt.in); got != tt.want {
				t.Errorf("collapseWhitespace(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
