package telegraphclient

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
)

const (
	titleLimit      = 256
	maxContentBytes = 60 * 1024
	timeLayout      = "2006-01-02 15:04 MST"
)

type node struct {
	Tag      string `json:"tag,omitempty"`
	Attrs    *attrs `json:"attrs,omitempty"`
	Children []any  `json:"children,omitempty"`
}

type attrs struct {
	Href string `json:"href,omitempty"`
	Src  string `json:"src,omitempty"`
}

type page struct {
	title   string
	content []any
}

func buildPages(inc incident.Incident, events []event.Event, maxBytes int) []page {
	header := headerNodes(inc)

	var pages []page
	content := append([]any{}, header...)
	size := nodesSize(content)

	flush := func() {
		pages = append(pages, page{content: content})
		content = append([]any{}, header...)
		size = nodesSize(header)
	}

	for _, e := range events {
		nodes := eventNodes(e)
		ns := nodesSize(nodes)
		if len(content) > len(header) && size+ns > maxBytes {
			flush()
		}
		content = append(content, nodes...)
		size += ns
	}
	pages = append(pages, page{content: content})

	titlePages(pages, inc.Title)
	return pages
}

func paginate(content []any, urls []string, current int) []any {
	if len(urls) < 2 {
		return content
	}

	nav := navNode(urls, current)
	out := make([]any, 0, len(content)+2)
	out = append(out, nav)
	out = append(out, content...)
	out = append(out, nav)
	return out
}

func navNode(urls []string, current int) node {
	children := []any{element("b", text("Pages: "))}
	for i, u := range urls {
		label := fmt.Sprintf("%d", i+1)
		if i == current {
			children = append(children, element("b", text(label)))
		} else {
			children = append(children, link(u, label))
		}
		if i < len(urls)-1 {
			children = append(children, text(" · "))
		}
	}
	return node{Tag: "p", Children: children}
}

func link(href, label string) node {
	return node{Tag: "a", Attrs: &attrs{Href: href}, Children: []any{text(label)}}
}

func titlePages(pages []page, title string) {
	total := len(pages)
	for i := range pages {
		t := title
		if total > 1 {
			t = fmt.Sprintf("%s (%d/%d)", title, i+1, total)
		}
		pages[i].title = truncate(t, titleLimit)
	}
}

func headerNodes(inc incident.Incident) []any {
	meta := fmt.Sprintf("Severity: %s · Status: %s · Opened: %s",
		inc.Severity, inc.Status, formatTime(inc.CreatedAt))
	if inc.ClosedAt != nil {
		meta += " · Closed: " + formatTime(*inc.ClosedAt)
	}

	return []any{
		element("h3", text("Timeline")),
		element("p", element("i", text(meta))),
	}
}

func eventNodes(e event.Event) []any {
	who := e.Username
	if who == "" {
		who = "system"
	}

	children := []any{
		element("b", text(formatTime(e.CreatedAt))),
		text(fmt.Sprintf(" — %s", who)),
	}
	if summary := e.Summary(); summary != "" {
		children = append(children, text(": "+summary))
	}

	nodes := []any{node{Tag: "p", Children: children}}
	if e.MediaURL != nil && *e.MediaURL != "" {
		nodes = append(nodes, imageNode(*e.MediaURL))
	}
	return nodes
}

func imageNode(src string) node {
	img := node{Tag: "img", Attrs: &attrs{Src: src}}
	return node{Tag: "figure", Children: []any{img}}
}

func element(tag string, children ...any) node {
	return node{Tag: tag, Children: children}
}

func text(s string) string { return s }

func formatTime(t time.Time) string { return t.Format(timeLayout) }

func truncate(s string, limit int) string {
	r := []rune(s)
	if len(r) <= limit {
		return s
	}
	return string(r[:limit])
}

func nodesSize(nodes []any) int {
	b, err := json.Marshal(nodes)
	if err != nil {
		return 0
	}
	return len(b)
}
