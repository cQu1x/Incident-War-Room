package telegraphclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/cQu1x/Incident-War-Room/internal/domain/timeline"
	"github.com/cQu1x/Incident-War-Room/internal/errs"
)

const (
	defaultBaseURL   = "https://api.telegra.ph"
	defaultShortName = "Incident War Room"
)

type Client struct {
	baseURL    string
	authorName string
	http       *http.Client

	mu    sync.Mutex
	token string
}

type Option func(*Client)

func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.http = h }
}

func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(baseURL, "/") }
}

func WithAccessToken(token string) Option {
	return func(c *Client) { c.token = strings.TrimSpace(token) }
}

func WithAuthorName(name string) Option {
	return func(c *Client) { c.authorName = name }
}

func New(opts ...Option) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		authorName: defaultShortName,
		http:       &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) Publish(ctx context.Context, t timeline.Timeline) ([]string, error) {
	const op = "telegraphclient.Publish"

	token, err := c.ensureToken(ctx)
	if err != nil {
		return nil, err
	}

	pages := buildPages(t.Incident, t.Events, maxContentBytes)
	existing := existingPages(t.Incident.TelegraphURLs)

	created := make([]createdPage, len(pages))
	for i, p := range pages {
		if i < len(existing) {
			created[i] = existing[i]
			continue
		}
		cp, err := c.createPage(ctx, token, p)
		if err != nil {
			return nil, errs.Wrapf(errs.KindUnavailable, op, err, "create telegraph page")
		}
		created[i] = cp
	}

	urls := make([]string, len(created))
	for i, cp := range created {
		urls[i] = cp.url
	}

	for i, p := range pages {
		reused := i < len(existing)
		if !reused && len(pages) == 1 {
			continue
		}
		content := p.content
		if len(pages) > 1 {
			content = paginate(p.content, urls, i)
		}
		if err := c.editPage(ctx, token, created[i].path, p.title, content); err != nil {
			return nil, errs.Wrapf(errs.KindUnavailable, op, err, "update telegraph page")
		}
	}

	return urls, nil
}

// existingPages recovers the Telegraph page paths from previously published
// URLs so the same pages can be edited in place instead of new ones being
// created on every publish. A Telegraph URL is https://telegra.ph/<path>, so the
// path is its last segment.
func existingPages(urls []string) []createdPage {
	pages := make([]createdPage, 0, len(urls))
	for _, u := range urls {
		path := u
		if i := strings.LastIndex(u, "/"); i >= 0 {
			path = u[i+1:]
		}
		if path == "" {
			continue
		}
		pages = append(pages, createdPage{url: u, path: path})
	}
	return pages
}

func (c *Client) ensureToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" {
		return c.token, nil
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := c.call(ctx, "createAccount", url.Values{
		"short_name":  {defaultShortName},
		"author_name": {c.authorName},
	}, &result); err != nil {
		return "", errs.Wrapf(errs.KindUnavailable, "telegraphclient.ensureToken", err, "create telegraph account")
	}

	c.token = result.AccessToken
	return c.token, nil
}

type createdPage struct {
	url  string
	path string
}

func (c *Client) createPage(ctx context.Context, token string, p page) (createdPage, error) {
	content, err := json.Marshal(p.content)
	if err != nil {
		return createdPage{}, errs.Wrapf(errs.KindInternal, "telegraphclient.createPage", err, "marshal content")
	}

	var result struct {
		URL  string `json:"url"`
		Path string `json:"path"`
	}
	if err := c.call(ctx, "createPage", url.Values{
		"access_token": {token},
		"title":        {p.title},
		"author_name":  {c.authorName},
		"content":      {string(content)},
	}, &result); err != nil {
		return createdPage{}, err
	}

	return createdPage{url: result.URL, path: result.Path}, nil
}

func (c *Client) editPage(ctx context.Context, token, path, title string, content []any) error {
	body, err := json.Marshal(content)
	if err != nil {
		return errs.Wrapf(errs.KindInternal, "telegraphclient.editPage", err, "marshal content")
	}

	return c.call(ctx, "editPage/"+path, url.Values{
		"access_token": {token},
		"title":        {title},
		"author_name":  {c.authorName},
		"content":      {string(body)},
	}, nil)
}

func (c *Client) call(ctx context.Context, method string, form url.Values, out any) error {
	const op = "telegraphclient.call"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/"+method, strings.NewReader(form.Encode()))
	if err != nil {
		return errs.Wrapf(errs.KindInternal, op, err, "build request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return errs.Wrapf(errs.KindUnavailable, op, err, "call telegraph")
	}
	defer resp.Body.Close()

	var envelope struct {
		OK     bool            `json:"ok"`
		Error  string          `json:"error"`
		Result json.RawMessage `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return errs.Wrapf(errs.KindUnavailable, op, err, "decode telegraph response")
	}
	if !envelope.OK {
		return errs.New(errs.KindUnavailable, op, "telegraph error: "+envelope.Error)
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(envelope.Result, out); err != nil {
		return errs.Wrapf(errs.KindInternal, op, err, "decode telegraph result")
	}
	return nil
}
