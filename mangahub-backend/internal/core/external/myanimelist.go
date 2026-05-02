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

// MyAnimeListClient calls the MAL official v2 REST API
// (https://api.myanimelist.net/v2). Authenticated requests need a registered
// X-MAL-CLIENT-ID; when ClientID is empty the client is effectively disabled.
type MyAnimeListClient struct {
	BaseURL    string
	HTTPClient *http.Client
	ClientID   string
	limiter    *rateLimiter
}

func NewMyAnimeListClient(baseURL, clientID string) *MyAnimeListClient {
	if baseURL == "" {
		baseURL = "https://api.myanimelist.net/v2"
	}
	return &MyAnimeListClient{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		ClientID:   clientID,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
		limiter:    newRateLimiter(3),
	}
}

const malFields = "alternative_titles,main_picture,synopsis,mean,popularity," +
	"num_chapters,status,genres,authors{first_name,last_name},updated_at"

func (c *MyAnimeListClient) SearchManga(ctx context.Context, q string, page int) ([]MALEntity, error) {
	if page <= 0 {
		page = 1
	}
	const limit = 20
	offset := (page - 1) * limit

	qs := url.Values{}
	qs.Set("q", q)
	qs.Set("limit", strconv.Itoa(limit))
	qs.Set("offset", strconv.Itoa(offset))
	qs.Set("fields", malFields)

	var raw malListResp
	if err := c.getJSON(ctx, "/manga?"+qs.Encode(), &raw); err != nil {
		return nil, err
	}
	out := make([]MALEntity, 0, len(raw.Data))
	for _, n := range raw.Data {
		out = append(out, malToEntity(n.Node))
	}
	return out, nil
}

func (c *MyAnimeListClient) GetManga(ctx context.Context, id string) (*MALEntity, error) {
	qs := url.Values{}
	qs.Set("fields", malFields)

	var raw malNode
	if err := c.getJSON(ctx, "/manga/"+url.PathEscape(id)+"?"+qs.Encode(), &raw); err != nil {
		return nil, err
	}
	e := malToEntity(raw)
	return &e, nil
}

func (c *MyAnimeListClient) getJSON(ctx context.Context, path string, out any) error {
	build := func() (*http.Request, error) {
		req, err := http.NewRequest(http.MethodGet, c.BaseURL+path, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		if c.ClientID != "" {
			req.Header.Set("X-MAL-CLIENT-ID", c.ClientID)
		}
		return req, nil
	}
	_, body, err := doWithRetry(ctx, c.HTTPClient, build, c.limiter, 3)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("mal decode: %w", err)
	}
	return nil
}

// --- Raw MAL response shapes ---

type malListResp struct {
	Data []struct {
		Node malNode `json:"node"`
	} `json:"data"`
}

type malNode struct {
	ID          int          `json:"id"`
	Title       string       `json:"title"`
	AltTitles   malAltTitles `json:"alternative_titles"`
	MainPicture struct {
		Large  string `json:"large"`
		Medium string `json:"medium"`
	} `json:"main_picture"`
	Synopsis    string      `json:"synopsis"`
	Mean        float64     `json:"mean"`
	Popularity  int         `json:"popularity"`
	NumChapters int         `json:"num_chapters"`
	Status      string      `json:"status"`
	Genres      []malName   `json:"genres"`
	Authors     []malAuthor `json:"authors"`
	UpdatedAt   string      `json:"updated_at"`
}

type malAltTitles struct {
	Synonyms []string `json:"synonyms"`
	EN       string   `json:"en"`
	JA       string   `json:"ja"`
}

type malName struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type malAuthor struct {
	Node struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	} `json:"node"`
	Role string `json:"role"`
}

// --- Mapping into SourceEntity ---

func malToEntity(n malNode) MALEntity {
	e := MALEntity{
		Source:      "myanimelist",
		ExternalID:  strconv.Itoa(n.ID),
		Title:       n.Title,
		Description: n.Synopsis,
		Status:      malStatus(n.Status),
		Chapters:    n.NumChapters,
		Rating:      n.Mean, // already on the 0–10 scale
		Popularity:  rankToPopularity(n.Popularity),
		CoverURL:    n.MainPicture.Large,
	}
	if n.AltTitles.EN != "" && n.AltTitles.EN != n.Title {
		e.AltTitles = append(e.AltTitles, n.AltTitles.EN)
	}
	if n.AltTitles.JA != "" && n.AltTitles.JA != n.Title {
		e.AltTitles = append(e.AltTitles, n.AltTitles.JA)
	}
	e.AltTitles = append(e.AltTitles, n.AltTitles.Synonyms...)
	for _, g := range n.Genres {
		e.Genres = append(e.Genres, g.Name)
	}
	for _, a := range n.Authors {
		full := strings.TrimSpace(a.Node.FirstName + " " + a.Node.LastName)
		if full == "" {
			continue
		}
		if strings.Contains(strings.ToLower(a.Role), "art") {
			e.Artists = append(e.Artists, full)
		} else {
			e.Authors = append(e.Authors, full)
		}
	}
	if t, err := time.Parse(time.RFC3339, n.UpdatedAt); err == nil {
		e.UpdatedAt = t
	}
	return e
}

func malStatus(s string) string {
	switch strings.ToLower(s) {
	case "currently_publishing":
		return "ongoing"
	case "finished":
		return "completed"
	case "on_hiatus":
		return "hiatus"
	case "discontinued":
		return "completed"
	default:
		return "ongoing"
	}
}

// MAL "popularity" is a rank where 1 = most popular. Invert so that "higher
// = more popular" matches the rest of the system.
func rankToPopularity(rank int) int {
	if rank <= 0 {
		return 0
	}
	const ceiling = 100000
	p := ceiling - rank
	if p < 0 {
		return 0
	}
	return p
}
