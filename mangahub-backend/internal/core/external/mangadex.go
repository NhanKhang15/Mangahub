package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// MangaDexClient talks to the public MangaDex REST API
// (https://api.mangadex.org). It enforces ~5 req/s by default and retries
// 3× with exponential backoff on 429/5xx.
type MangaDexClient struct {
	BaseURL    string
	HTTPClient *http.Client
	AuthToken  string
	limiter    *rateLimiter
}

func NewMangaDexClient(baseURL, token string) *MangaDexClient {
	if baseURL == "" {
		baseURL = "https://api.mangadex.org"
	}
	return &MangaDexClient{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		AuthToken:  token,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
		limiter:    newRateLimiter(5),
	}
}

// SearchManga returns up to 20 results from /manga?title=…&offset=…&limit=20.
func (c *MangaDexClient) SearchManga(ctx context.Context, q string, page int) ([]MangaDexEntity, error) {
	if page <= 0 {
		page = 1
	}
	const limit = 20
	offset := (page - 1) * limit

	qs := url.Values{}
	if q != "" {
		qs.Set("title", q)
	}
	qs.Set("limit", strconv.Itoa(limit))
	qs.Set("offset", strconv.Itoa(offset))
	qs.Set("order[followedCount]", "desc")
	qs.Add("includes[]", "author")
	qs.Add("includes[]", "artist")
	qs.Add("includes[]", "cover_art")

	var raw mdListResp
	if err := c.getJSON(ctx, "/manga?"+qs.Encode(), &raw); err != nil {
		return nil, err
	}
	out := make([]MangaDexEntity, 0, len(raw.Data))
	for _, m := range raw.Data {
		out = append(out, mdToEntity(m))
	}
	return out, nil
}

// GetManga returns the manga with the given MangaDex UUID.
func (c *MangaDexClient) GetManga(ctx context.Context, id string) (*MangaDexEntity, error) {
	path := "/manga/" + url.PathEscape(id) +
		"?includes[]=author&includes[]=artist&includes[]=cover_art"
	var raw mdSingleResp
	if err := c.getJSON(ctx, path, &raw); err != nil {
		return nil, err
	}
	e := mdToEntity(raw.Data)
	return &e, nil
}

// GetChapters returns up to 100 English chapters of the given manga, ordered
// by chapter number ascending.
func (c *MangaDexClient) GetChapters(ctx context.Context, id string) ([]Chapter, error) {
	qs := url.Values{}
	qs.Set("manga", id)
	qs.Set("limit", "100")
	qs.Set("translatedLanguage[]", "en")
	qs.Set("order[chapter]", "asc")

	var raw mdChapterListResp
	if err := c.getJSON(ctx, "/chapter?"+qs.Encode(), &raw); err != nil {
		return nil, err
	}
	out := make([]Chapter, 0, len(raw.Data))
	for _, ch := range raw.Data {
		t, _ := time.Parse(time.RFC3339, ch.Attributes.PublishAt)
		out = append(out, Chapter{
			ID:        ch.ID,
			MangaID:   id,
			Number:    ch.Attributes.Chapter,
			Title:     ch.Attributes.Title,
			Language:  ch.Attributes.Language,
			PublishAt: t,
		})
	}
	return out, nil
}

func (c *MangaDexClient) getJSON(ctx context.Context, path string, out any) error {
	build := func() (*http.Request, error) {
		req, err := http.NewRequest(http.MethodGet, c.BaseURL+path, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		if c.AuthToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.AuthToken)
		}
		return req, nil
	}
	_, body, err := doWithRetry(ctx, c.HTTPClient, build, c.limiter, 3)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("mangadex decode: %w", err)
	}
	return nil
}

// --- Raw MangaDex response shapes ---

type mdListResp struct {
	Result string    `json:"result"`
	Data   []mdManga `json:"data"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
	Total  int       `json:"total"`
}

type mdSingleResp struct {
	Result string  `json:"result"`
	Data   mdManga `json:"data"`
}

type mdManga struct {
	ID            string  `json:"id"`
	Type          string  `json:"type"`
	Attributes    mdAttrs `json:"attributes"`
	Relationships []mdRel `json:"relationships"`
}

type mdAttrs struct {
	Title       map[string]string   `json:"title"`
	AltTitles   []map[string]string `json:"altTitles"`
	Description map[string]string   `json:"description"`
	Status      string              `json:"status"`
	Year        int                 `json:"year"`
	LastChapter string              `json:"lastChapter"`
	UpdatedAt   string              `json:"updatedAt"`
	Tags        []mdTag             `json:"tags"`
}

type mdTag struct {
	Attributes struct {
		Name  map[string]string `json:"name"`
		Group string            `json:"group"`
	} `json:"attributes"`
}

type mdRel struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Attributes json.RawMessage `json:"attributes"`
}

type mdChapterListResp struct {
	Result string      `json:"result"`
	Data   []mdChapter `json:"data"`
}

type mdChapter struct {
	ID         string         `json:"id"`
	Attributes mdChapterAttrs `json:"attributes"`
}

type mdChapterAttrs struct {
	Volume    string `json:"volume"`
	Chapter   string `json:"chapter"`
	Title     string `json:"title"`
	Language  string `json:"translatedLanguage"`
	PublishAt string `json:"publishAt"`
}

// --- Mapping into the canonical SourceEntity ---

func mdToEntity(m mdManga) MangaDexEntity {
	e := MangaDexEntity{
		Source:      "mangadex",
		ExternalID:  m.ID,
		Title:       pickFirst(m.Attributes.Title, "en"),
		Description: pickFirst(m.Attributes.Description, "en"),
		Status:      normalizeStatus(m.Attributes.Status),
	}
	for _, alt := range m.Attributes.AltTitles {
		if v := pickFirst(alt, "en"); v != "" {
			e.AltTitles = append(e.AltTitles, v)
		}
	}
	for _, t := range m.Attributes.Tags {
		name := pickFirst(t.Attributes.Name, "en")
		if name == "" {
			continue
		}
		if t.Attributes.Group == "genre" {
			e.Genres = append(e.Genres, name)
		} else {
			e.Tags = append(e.Tags, name)
		}
	}
	if n, err := strconv.Atoi(m.Attributes.LastChapter); err == nil {
		e.Chapters = n
	}
	if t, err := time.Parse(time.RFC3339, m.Attributes.UpdatedAt); err == nil {
		e.UpdatedAt = t
	}
	for _, rel := range m.Relationships {
		switch rel.Type {
		case "author":
			if name := relName(rel); name != "" {
				e.Authors = append(e.Authors, name)
			}
		case "artist":
			if name := relName(rel); name != "" {
				e.Artists = append(e.Artists, name)
			}
		case "cover_art":
			if filename := relCoverFile(rel); filename != "" {
				e.CoverURL = "https://uploads.mangadex.org/covers/" + m.ID + "/" + filename
			}
		}
	}
	return e
}

func relName(rel mdRel) string {
	if len(rel.Attributes) == 0 {
		return ""
	}
	var attrs struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal(rel.Attributes, &attrs)
	return attrs.Name
}

func relCoverFile(rel mdRel) string {
	if len(rel.Attributes) == 0 {
		return ""
	}
	var attrs struct {
		FileName string `json:"fileName"`
	}
	_ = json.Unmarshal(rel.Attributes, &attrs)
	return attrs.FileName
}

func pickFirst(m map[string]string, lang string) string {
	if v, ok := m[lang]; ok && v != "" {
		return v
	}
	for _, v := range m {
		if v != "" {
			return v
		}
	}
	return ""
}

func normalizeStatus(s string) string {
	switch strings.ToLower(s) {
	case "ongoing":
		return "ongoing"
	case "completed":
		return "completed"
	case "hiatus":
		return "hiatus"
	case "cancelled", "canceled":
		return "completed"
	default:
		return "ongoing"
	}
}
