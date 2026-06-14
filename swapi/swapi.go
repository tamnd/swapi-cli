// Package swapi is the library behind the swapi command line:
// the HTTP client, request shaping, and typed data models for the Star Wars API.
//
// The Client sets a real User-Agent, paces requests, and retries the transient
// failures (429 and 5xx) that any public API throws under load. All six resource
// types are supported: people, films, planets, starships, vehicles, and species.
package swapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Host is the site this client talks to.
const Host = "swapi.py4e.com"

// BaseURL is the root every request is built from.
const BaseURL = "https://" + Host

// Config holds tunable knobs for the HTTP client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns sensible defaults for production use.
func DefaultConfig() Config {
	return Config{
		BaseURL:   BaseURL,
		UserAgent: "swapi-cli/0.1.0 (github.com/tamnd/swapi-cli)",
		Rate:      200 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
	}
}

// Client talks to the Star Wars API over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// envelope is the paginated list wrapper all SWAPI list endpoints return.
type envelope[T any] struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []T    `json:"results"`
}

// People returns Star Wars characters. search is optional ("" means all).
// Follows pagination until limit is satisfied or no more pages remain.
func (c *Client) People(ctx context.Context, search string, limit int) ([]Person, error) {
	u := c.cfg.BaseURL + "/api/people/"
	if search != "" {
		u += "?search=" + url.QueryEscape(search)
	}
	return listPages[Person](ctx, c, u, limit)
}

// Films returns all 7 canonical Star Wars films.
func (c *Client) Films(ctx context.Context) ([]Film, error) {
	b, err := c.get(ctx, c.cfg.BaseURL+"/api/films/")
	if err != nil {
		return nil, err
	}
	var env envelope[Film]
	if err := json.Unmarshal(b, &env); err != nil {
		return nil, fmt.Errorf("decode films: %w", err)
	}
	return env.Results, nil
}

// Planets returns planets in the Star Wars universe. search is optional.
func (c *Client) Planets(ctx context.Context, search string, limit int) ([]Planet, error) {
	u := c.cfg.BaseURL + "/api/planets/"
	if search != "" {
		u += "?search=" + url.QueryEscape(search)
	}
	return listPages[Planet](ctx, c, u, limit)
}

// Starships returns starships used in the Star Wars films. search is optional.
func (c *Client) Starships(ctx context.Context, search string, limit int) ([]Starship, error) {
	u := c.cfg.BaseURL + "/api/starships/"
	if search != "" {
		u += "?search=" + url.QueryEscape(search)
	}
	return listPages[Starship](ctx, c, u, limit)
}

// Vehicles returns vehicles from the Star Wars films. search is optional.
func (c *Client) Vehicles(ctx context.Context, search string, limit int) ([]Vehicle, error) {
	u := c.cfg.BaseURL + "/api/vehicles/"
	if search != "" {
		u += "?search=" + url.QueryEscape(search)
	}
	return listPages[Vehicle](ctx, c, u, limit)
}

// Species returns species/races from the Star Wars universe. search is optional.
func (c *Client) Species(ctx context.Context, search string, limit int) ([]Species, error) {
	u := c.cfg.BaseURL + "/api/species/"
	if search != "" {
		u += "?search=" + url.QueryEscape(search)
	}
	return listPages[Species](ctx, c, u, limit)
}

// listPages walks SWAPI pagination until limit items are collected or next is empty.
func listPages[T any](ctx context.Context, c *Client, startURL string, limit int) ([]T, error) {
	var out []T
	urlStr := startURL
	for urlStr != "" {
		b, err := c.get(ctx, urlStr)
		if err != nil {
			return nil, err
		}
		var env envelope[T]
		if err := json.Unmarshal(b, &env); err != nil {
			return nil, fmt.Errorf("decode page: %w", err)
		}
		out = append(out, env.Results...)
		if limit > 0 && len(out) >= limit {
			out = out[:limit]
			break
		}
		urlStr = env.Next
	}
	return out, nil
}

// get fetches url and returns the response body. It paces and retries according
// to the client's settings.
func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

// pace blocks until at least Rate has passed since the previous request.
func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
