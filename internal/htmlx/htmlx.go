// Package htmlx
package htmlx

import (
	"strings"

	"golang.org/x/net/html"
)

// HTMLToText extracts visible text from an HTML fragment, dropping the
// contents of <script> and <style> tags. Returns whitespace collapsed text.
func HTMLToText(htmlStr string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	var walk func(*html.Node)

	// Insert breaks at block level boundaries
	blockTags := map[string]bool{
		"p": true, "div": true, "br": true, "li": true, "ul": true, "ol": true,
		"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
		"blockquote": true, "pre": true, "section": true, "article": true, "header": true, "footer": true, "hr": true,
		"table": true, "tr": true, "td": true, "th": true, "dd": true, "dt": true, "figure": true,
	}

	walk = func(n *html.Node) {
		// Skip the entire subtree of the noise tags
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}
		if n.Type == html.ElementNode && blockTags[n.Data] {
			sb.WriteString("\n")
		}
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	return collapseWhitespace(sb.String()), nil
}

func collapseWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		if trimmed := strings.Join(strings.Fields(line), " "); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return strings.Join(out, "\n")
}
