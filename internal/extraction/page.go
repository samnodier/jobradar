// Package extraction reads public page metadata without running a language model.
package extraction

import (
	"encoding/json"
	"github.com/samnodier/jobradar/internal/htmlx"
	"github.com/samnodier/jobradar/internal/llm"
	"golang.org/x/net/html"
	"strings"
)

// Page returns a reviewable draft. Structured JobPosting data takes precedence;
// a plain page fallback never guesses company, salary, skills or remote status.
func Page(raw string) (llm.ExtractionResult, bool) {
	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		return llm.ExtractionResult{}, false
	}
	var found map[string]any
	title := ""
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			title = n.FirstChild.Data
		}
		if n.Type == html.ElementNode && n.Data == "script" {
			for _, a := range n.Attr {
				if a.Key == "type" && strings.EqualFold(a.Val, "application/ld+json") && n.FirstChild != nil {
					var data any
					if json.Unmarshal([]byte(n.FirstChild.Data), &data) == nil && found == nil {
						found = findPosting(data)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	result := llm.ExtractionResult{Title: strings.TrimSpace(title), Skills: []string{}}
	if found != nil {
		result.Title = text(found["title"])
		if industry := text(found["industry"]); industry != "" {
			result.Industry = &industry
		}
		if category := text(found["occupationalCategory"]); category != "" {
			result.Category = &category
		}
		if kind := text(found["employmentType"]); kind != "" {
			kind = strings.ReplaceAll(strings.ToLower(kind), "_", "-")
			result.EmploymentType = &kind
		}
		result.Description = text(found["description"])
		if org, ok := found["hiringOrganization"].(map[string]any); ok {
			result.CompanyName = text(org["name"])
		}
		if strings.EqualFold(text(found["jobLocationType"]), "TELECOMMUTE") {
			v := true
			result.IsRemote = &v
		}
		locations := []string{}
		collectLocations(found["jobLocation"], &locations)
		// Remote postings can restrict applicant geography. Retain that information.
		if requirements, ok := found["applicantLocationRequirements"].([]any); ok {
			for _, v := range requirements {
				if m, ok := v.(map[string]any); ok {
					locations = append(locations, text(m["name"]))
				}
			}
		} else if m, ok := found["applicantLocationRequirements"].(map[string]any); ok {
			locations = append(locations, text(m["name"]))
		}
		location := strings.Join(locations, "; ")
		if location != "" {
			result.JobLocation = &location
		}
		if result.Title != "" && result.Description != "" {
			return result, true
		}
	}
	visible, _ := htmlx.HTMLToText(raw)
	result.Description = llm.LimitText(visible, 24000)
	return result, false
}
func text(v any) string { s, _ := v.(string); return strings.TrimSpace(s) }
func findPosting(v any) map[string]any {
	switch m := v.(type) {
	case map[string]any:
		typ := text(m["@type"])
		if typ == "JobPosting" {
			return m
		}
		if types, ok := m["@type"].([]any); ok {
			for _, t := range types {
				if text(t) == "JobPosting" {
					return m
				}
			}
		}
		for _, child := range m {
			if match := findPosting(child); match != nil {
				return match
			}
		}
	case []any:
		for _, child := range m {
			if match := findPosting(child); match != nil {
				return match
			}
		}
	}
	return nil
}
func collectLocations(v any, out *[]string) {
	if items, ok := v.([]any); ok {
		for _, item := range items {
			collectLocations(item, out)
		}
		return
	}
	m, ok := v.(map[string]any)
	if !ok {
		return
	}
	address, ok := m["address"].(map[string]any)
	if !ok {
		return
	}
	parts := []string{}
	for _, key := range []string{"addressLocality", "addressRegion", "addressCountry"} {
		value := text(address[key])
		if object, ok := address[key].(map[string]any); ok {
			value = text(object["name"])
		}
		if value != "" {
			parts = append(parts, value)
		}
	}
	if len(parts) > 0 {
		*out = append(*out, strings.Join(parts, ", "))
	}
}
