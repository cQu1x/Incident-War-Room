package telegraphclient

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/cQu1x/Incident-War-Room/internal/domain/event"
	"github.com/cQu1x/Incident-War-Room/internal/domain/incident"
)

func makeEvents(n int) []event.Event {
	events := make([]event.Event, n)
	base := time.Date(2026, 6, 20, 10, 0, 0, 0, time.UTC)
	for i := range events {
		events[i] = event.Event{
			Username:  "alice",
			Message:   strings.Repeat("x", 200),
			CreatedAt: base.Add(time.Duration(i) * time.Minute),
		}
	}
	return events
}

func TestBuildPagesSinglePage(t *testing.T) {
	pages := buildPages(incident.Incident{Title: "outage"}, makeEvents(3), maxContentBytes)

	if len(pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(pages))
	}
	if pages[0].title != "outage" {
		t.Fatalf("unexpected title %q", pages[0].title)
	}
}

func TestBuildPagesSplitsWhenOverflowing(t *testing.T) {
	pages := buildPages(incident.Incident{Title: "outage"}, makeEvents(20), 1024)

	if len(pages) < 2 {
		t.Fatalf("expected the timeline to split across pages, got %d", len(pages))
	}
	for i, p := range pages {
		if !strings.Contains(p.title, "outage") {
			t.Errorf("page %d title %q missing incident title", i, p.title)
		}
		if !strings.Contains(p.title, "/") {
			t.Errorf("page %d title %q missing part suffix", i, p.title)
		}
	}
}

func TestBuildPagesKeepsLongEventOnItsOwnPage(t *testing.T) {
	huge := event.Event{Username: "alice", Message: strings.Repeat("y", 4000), CreatedAt: time.Now()}
	events := []event.Event{huge, huge, huge}

	pages := buildPages(incident.Incident{Title: "outage"}, events, 1024)
	if len(pages) != 3 {
		t.Fatalf("expected each oversized event on its own page, got %d pages", len(pages))
	}
}

func TestBuildPagesLabelsLifecycleEvents(t *testing.T) {
	events := []event.Event{
		{Type: event.TypeIncidentCreated, Username: "alice", Message: "outage", CreatedAt: time.Now()},
		{Type: event.TypeIncidentClosed, Username: "bob", CreatedAt: time.Now()},
		{Type: event.TypeIncidentReopened, Username: "carol", CreatedAt: time.Now()},
	}

	pages := buildPages(incident.Incident{Title: "outage"}, events, maxContentBytes)
	raw, _ := json.Marshal(pages[0].content)
	s := string(raw)

	for _, want := range []string{"Incident opened: outage", "Incident closed", "Incident reopened"} {
		if !strings.Contains(s, want) {
			t.Errorf("page content missing lifecycle label %q: %s", want, s)
		}
	}
}

func TestBuildPagesRendersEventImage(t *testing.T) {
	url := "https://example.com/photo.jpg"
	events := []event.Event{{Username: "alice", Message: "see attached", MediaURLs: []string{url}, CreatedAt: time.Now()}}

	pages := buildPages(incident.Incident{Title: "outage"}, events, maxContentBytes)

	raw, _ := json.Marshal(pages[0].content)
	s := string(raw)
	if !strings.Contains(s, `"tag":"img"`) {
		t.Errorf("expected an img node, got %s", s)
	}
	if !strings.Contains(s, url) {
		t.Errorf("img node missing media url: %s", s)
	}
}

func TestBuildPagesSkipsEmptyMediaURL(t *testing.T) {
	events := []event.Event{{Username: "alice", Message: "no photo", MediaURLs: []string{""}, CreatedAt: time.Now()}}

	pages := buildPages(incident.Incident{Title: "outage"}, events, maxContentBytes)

	raw, _ := json.Marshal(pages[0].content)
	if strings.Contains(string(raw), `"tag":"img"`) {
		t.Errorf("expected no img node for empty media url: %s", raw)
	}
}

func TestBuildPagesRendersNonImageMediaAsLink(t *testing.T) {
	url := "https://example.com/capture.mp4"
	events := []event.Event{{Username: "alice", Message: "clip", MediaURLs: []string{url}, CreatedAt: time.Now()}}

	pages := buildPages(incident.Incident{Title: "outage"}, events, maxContentBytes)

	raw, _ := json.Marshal(pages[0].content)
	s := string(raw)
	if strings.Contains(s, `"tag":"img"`) {
		t.Errorf("non-image media should not render as an img node: %s", s)
	}
	if !strings.Contains(s, `"tag":"a"`) || !strings.Contains(s, url) {
		t.Errorf("expected a link to the attachment: %s", s)
	}
}

func TestPaginateAddsNavLinks(t *testing.T) {
	content := []any{element("p", text("body"))}
	urls := []string{"https://telegra.ph/a", "https://telegra.ph/b", "https://telegra.ph/c"}

	got := paginate(content, urls, 1)

	if len(got) != len(content)+2 {
		t.Fatalf("expected nav prepended and appended, got %d nodes", len(got))
	}

	raw, _ := json.Marshal(got)
	s := string(raw)
	for _, u := range []string{urls[0], urls[2]} {
		if !strings.Contains(s, u) {
			t.Errorf("nav missing link to %q: %s", u, s)
		}
	}
	if strings.Count(s, urls[1]) != 0 {
		t.Errorf("current page should not link to itself: %s", s)
	}
}

func TestPaginateNoopForSinglePage(t *testing.T) {
	content := []any{element("p", text("body"))}
	got := paginate(content, []string{"https://telegra.ph/a"}, 0)

	if len(got) != len(content) {
		t.Fatalf("single page should not get nav, got %d nodes", len(got))
	}
}
