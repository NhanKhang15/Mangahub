package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// AniListClient talks to the AniList GraphQL endpoint
// (https://graphql.anilist.co). All operations go through postGraphQL —
// the single Post-JSON helper required by the Phase 3 plan.
type AniListClient struct {
	BaseURL    string
	HTTPClient *http.Client
	AuthToken  string
	limiter    *rateLimiter
}

func NewAniListClient(baseURL, token string) *AniListClient {
	if baseURL == "" {
		baseURL = "https://graphql.anilist.co"
	}
	return &AniListClient{
		BaseURL:    baseURL,
		AuthToken:  token,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
		// AniList allows ~90 req/min; round down to be polite.
		limiter: newRateLimiter(2),
	}
}

const alMediaSelection = `
  id
  title { romaji english native }
  synonyms
  description(asHtml: false)
  averageScore
  popularity
  chapters
  status
  genres
  tags { name }
  coverImage { large }
  updatedAt
  staff { edges { role node { name { full } } } }
`

var (
	anilistSearchQuery = `query ($q: String, $page: Int, $perPage: Int) {
  Page(page: $page, perPage: $perPage) {
    media(search: $q, type: MANGA, sort: POPULARITY_DESC) {` + alMediaSelection + `}
  }
}`

	anilistTrendingQuery = `query ($page: Int, $perPage: Int) {
  Page(page: $page, perPage: $perPage) {
    media(type: MANGA, sort: TRENDING_DESC) {` + alMediaSelection + `}
  }
}`

	anilistGetQuery = `query ($id: Int) {
  Media(id: $id, type: MANGA) {` + alMediaSelection + `}
}`
)

func (c *AniListClient) SearchManga(ctx context.Context, q string, page int) ([]AniListEntity, error) {
	if page <= 0 {
		page = 1
	}
	media, err := c.fetchPage(ctx, anilistSearchQuery, map[string]any{
		"q":       q,
		"page":    page,
		"perPage": 20,
	})
	if err != nil {
		return nil, err
	}
	return mapMedia(media), nil
}

// Trending returns the top trending manga in a single page (used by the
// seed script). perPage is capped by AniList at 50, so larger N must paginate.
func (c *AniListClient) Trending(ctx context.Context, page, perPage int) ([]AniListEntity, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 50 {
		perPage = 50
	}
	media, err := c.fetchPage(ctx, anilistTrendingQuery, map[string]any{
		"page":    page,
		"perPage": perPage,
	})
	if err != nil {
		return nil, err
	}
	return mapMedia(media), nil
}

func (c *AniListClient) GetManga(ctx context.Context, id string) (*AniListEntity, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("anilist id must be int: %w", err)
	}
	var resp struct {
		Data struct {
			Media alMedia `json:"Media"`
		} `json:"data"`
	}
	if err := c.postGraphQL(ctx, anilistGetQuery, map[string]any{"id": n}, &resp); err != nil {
		return nil, err
	}
	e := alToEntity(resp.Data.Media)
	return &e, nil
}

func (c *AniListClient) fetchPage(ctx context.Context, query string, vars map[string]any) ([]alMedia, error) {
	var resp struct {
		Data struct {
			Page struct {
				Media []alMedia `json:"media"`
			} `json:"Page"`
		} `json:"data"`
	}
	if err := c.postGraphQL(ctx, query, vars, &resp); err != nil {
		return nil, err
	}
	return resp.Data.Page.Media, nil
}

// postGraphQL is the single Post-JSON helper required by the Phase 3 plan.
// It sends {query, variables} as JSON, retries through doWithRetry, and
// surfaces GraphQL-level errors as a normal error.
func (c *AniListClient) postGraphQL(ctx context.Context, query string, vars map[string]any, out any) error {
	body, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": vars,
	})
	if err != nil {
		return err
	}
	build := func() (*http.Request, error) {
		req, err := http.NewRequest(http.MethodPost, c.BaseURL, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		if c.AuthToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.AuthToken)
		}
		return req, nil
	}
	_, respBody, err := doWithRetry(ctx, c.HTTPClient, build, c.limiter, 3)
	if err != nil {
		return err
	}
	var envelope struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(respBody, &envelope); err == nil && len(envelope.Errors) > 0 {
		return fmt.Errorf("anilist graphql error: %s", envelope.Errors[0].Message)
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("anilist decode: %w", err)
	}
	return nil
}

// --- Raw AniList response shapes ---

type alMedia struct {
	ID    int `json:"id"`
	Title struct {
		Romaji  string `json:"romaji"`
		English string `json:"english"`
		Native  string `json:"native"`
	} `json:"title"`
	Synonyms     []string `json:"synonyms"`
	Description  string   `json:"description"`
	AverageScore int      `json:"averageScore"`
	Popularity   int      `json:"popularity"`
	Chapters     int      `json:"chapters"`
	Status       string   `json:"status"`
	Genres       []string `json:"genres"`
	Tags         []struct {
		Name string `json:"name"`
	} `json:"tags"`
	CoverImage struct {
		Large string `json:"large"`
	} `json:"coverImage"`
	UpdatedAt int64 `json:"updatedAt"`
	Staff     struct {
		Edges []struct {
			Role string `json:"role"`
			Node struct {
				Name struct {
					Full string `json:"full"`
				} `json:"name"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"staff"`
}

// --- Mapping into SourceEntity ---

func mapMedia(in []alMedia) []AniListEntity {
	out := make([]AniListEntity, 0, len(in))
	for _, m := range in {
		out = append(out, alToEntity(m))
	}
	return out
}

func alToEntity(m alMedia) AniListEntity {
	title := m.Title.English
	if title == "" {
		title = m.Title.Romaji
	}
	if title == "" {
		title = m.Title.Native
	}

	e := AniListEntity{
		Source:      "anilist",
		ExternalID:  strconv.Itoa(m.ID),
		Title:       title,
		Description: stripHTML(m.Description),
		Status:      alStatus(m.Status),
		Chapters:    m.Chapters,
		Rating:      float64(m.AverageScore) / 10.0, // 0–100 → 0–10
		Popularity:  m.Popularity,
		CoverURL:    m.CoverImage.Large,
	}
	if m.Title.Romaji != "" && m.Title.Romaji != title {
		e.AltTitles = append(e.AltTitles, m.Title.Romaji)
	}
	if m.Title.Native != "" && m.Title.Native != title {
		e.AltTitles = append(e.AltTitles, m.Title.Native)
	}
	e.AltTitles = append(e.AltTitles, m.Synonyms...)
	e.Genres = append(e.Genres, m.Genres...)
	for _, t := range m.Tags {
		e.Tags = append(e.Tags, t.Name)
	}
	for _, edge := range m.Staff.Edges {
		full := strings.TrimSpace(edge.Node.Name.Full)
		if full == "" {
			continue
		}
		role := strings.ToLower(edge.Role)
		if strings.Contains(role, "art") {
			e.Artists = append(e.Artists, full)
		} else {
			e.Authors = append(e.Authors, full)
		}
	}
	if m.UpdatedAt > 0 {
		e.UpdatedAt = time.Unix(m.UpdatedAt, 0).UTC()
	}
	return e
}

func alStatus(s string) string {
	switch strings.ToUpper(s) {
	case "RELEASING":
		return "ongoing"
	case "FINISHED":
		return "completed"
	case "HIATUS":
		return "hiatus"
	case "CANCELLED":
		return "completed"
	default:
		return "ongoing"
	}
}

// stripHTML rips the simple <i>/<br>/<b> markup AniList ships in description.
func stripHTML(s string) string {
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")
	var b strings.Builder
	in := false
	for _, r := range s {
		switch r {
		case '<':
			in = true
		case '>':
			in = false
		default:
			if !in {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}
