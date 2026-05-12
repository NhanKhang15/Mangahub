// Package notifier ships gateway-side events to the standalone TCP and UDP
// servers via their internal HTTP publish APIs. It is intentionally
// fire-and-forget: failures are logged but never bubble up to the HTTP
// handler, so a downstream notifier outage cannot break the user-facing API.
package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

// Client posts JSON payloads to tcp-svc and udp-svc. Both target URLs are
// optional — an empty URL disables that channel.
type Client struct {
	tcpPublishURL string
	udpPublishURL string
	token         string
	http          *http.Client
}

func New(tcpURL, udpURL, token string) *Client {
	return &Client{
		tcpPublishURL: tcpURL,
		udpPublishURL: udpURL,
		token:         token,
		http: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

// ProgressEvent is the payload sent to tcp-svc.
type ProgressEvent struct {
	UserID     string `json:"user_id"`
	MangaID    string `json:"manga_id"`
	MangaTitle string `json:"manga_title,omitempty"`
	Chapter    int    `json:"chapter"`
	Status     string `json:"status"`
}

// ChapterEvent is the payload sent to udp-svc.
type ChapterEvent struct {
	MangaID    string `json:"manga_id"`
	MangaTitle string `json:"manga_title,omitempty"`
	Chapter    int    `json:"chapter"`
	Message    string `json:"message,omitempty"`
}

// PublishProgress pushes a ProgressEvent to tcp-svc. Safe to call in a
// goroutine; never returns an error to the caller.
func (c *Client) PublishProgress(ctx context.Context, ev ProgressEvent) {
	if c == nil || c.tcpPublishURL == "" {
		return
	}
	if err := c.post(ctx, c.tcpPublishURL, ev); err != nil {
		log.Printf("notifier: tcp publish failed: %v", err)
	}
}

// PublishChapter pushes a ChapterEvent to udp-svc. Same fire-and-forget
// semantics as PublishProgress.
func (c *Client) PublishChapter(ctx context.Context, ev ChapterEvent) {
	if c == nil || c.udpPublishURL == "" {
		return
	}
	if err := c.post(ctx, c.udpPublishURL, ev); err != nil {
		log.Printf("notifier: udp publish failed: %v", err)
	}
}

func (c *Client) post(ctx context.Context, url string, body any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("X-Internal-Token", c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return errors.New(resp.Status + ": " + string(body))
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}
