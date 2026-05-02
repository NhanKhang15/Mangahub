package external

import (
	"context"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	mangaModel "mangahub-backend/internal/modules/manga/model"
		)

// Aggregator fans a search query out to MangaDex / MAL / AniList in parallel
// (errgroup) and merges results into canonical mangaModel.Manga records — one per
// title cluster. Any client may be nil to disable that source.
type Aggregator struct {
	MangaDex *MangaDexClient
	MAL      *MyAnimeListClient
	AniList  *AniListClient
}

func NewAggregator(md *MangaDexClient, mal *MyAnimeListClient, al *AniListClient) *Aggregator {
	return &Aggregator{MangaDex: md, MAL: mal, AniList: al}
}

// Search hits all configured sources concurrently. If a single source errors,
// the entire call fails (errgroup semantics) so the caller can see upstream
// trouble rather than silently importing a partial set.
func (a *Aggregator) Search(ctx context.Context, q string, page int) ([]*mangaModel.Manga, error) {
	var (
		mdRes  []SourceEntity
		malRes []SourceEntity
		alRes  []SourceEntity
	)
	g, gctx := errgroup.WithContext(ctx)
	if a.MangaDex != nil {
		g.Go(func() error {
			r, err := a.MangaDex.SearchManga(gctx, q, page)
			if err != nil {
				return err
			}
			mdRes = r
			return nil
		})
	}
	if a.MAL != nil {
		g.Go(func() error {
			r, err := a.MAL.SearchManga(gctx, q, page)
			if err != nil {
				return err
			}
			malRes = r
			return nil
		})
	}
	if a.AniList != nil {
		g.Go(func() error {
			r, err := a.AniList.SearchManga(gctx, q, page)
			if err != nil {
				return err
			}
			alRes = r
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return MergeEntities(mdRes, malRes, alRes), nil
}

// Trending returns the top-N trending manga used by scripts/seed.go.
// AniList exposes a TRENDING_DESC sort so we use it as the source of truth;
// fall back to MangaDex's followedCount ordering if AniList is unavailable.
func (a *Aggregator) Trending(ctx context.Context, n int) ([]*mangaModel.Manga, error) {
	if n <= 0 {
		n = 100
	}
	if a.AniList != nil {
		var collected []SourceEntity
		const perPage = 50
		for page := 1; len(collected) < n; page++ {
			batch, err := a.AniList.Trending(ctx, page, perPage)
			if err != nil {
				return nil, err
			}
			if len(batch) == 0 {
				break
			}
			collected = append(collected, batch...)
		}
		if len(collected) > n {
			collected = collected[:n]
		}
		return MergeEntities(nil, nil, collected), nil
	}
	if a.MangaDex != nil {
		var collected []SourceEntity
		for page := 1; len(collected) < n; page++ {
			batch, err := a.MangaDex.SearchManga(ctx, "", page)
			if err != nil {
				return nil, err
			}
			if len(batch) == 0 {
				break
			}
			collected = append(collected, batch...)
		}
		if len(collected) > n {
			collected = collected[:n]
		}
		return MergeEntities(collected, nil, nil), nil
	}
	return nil, nil
}

// MergeEntities collapses entries from each source by title similarity. Each
// resulting cluster is materialized into one mangaModel.Manga whose external_ids
// map records every source we matched.
func MergeEntities(md, mal, al []SourceEntity) []*mangaModel.Manga {
	var clusters [][]SourceEntity
	add := func(ents []SourceEntity) {
		for _, e := range ents {
			placed := false
			for i, c := range clusters {
				if titleSimilar(c[0].Title, e.Title) {
					clusters[i] = append(clusters[i], e)
					placed = true
					break
				}
			}
			if !placed {
				clusters = append(clusters, []SourceEntity{e})
			}
		}
	}
	add(md)
	add(mal)
	add(al)

	out := make([]*mangaModel.Manga, 0, len(clusters))
	for _, c := range clusters {
		out = append(out, clusterToManga(c))
	}
	return out
}

func clusterToManga(c []SourceEntity) *mangaModel.Manga {
	m := &mangaModel.Manga{
		ExternalIDs: map[string]string{},
		Status:      mangaModel.MangaStatusOngoing,
	}
	altSet := map[string]bool{}
	genreSet := map[string]bool{}
	tagSet := map[string]bool{}
	var ratingSum float64
	var ratingN int

	for _, e := range c {
		if e.Source != "" && e.ExternalID != "" {
			m.ExternalIDs[e.Source] = e.ExternalID
		}

		// Pick the most reader-friendly title (English / ASCII) and demote
		// the rest into alt_titles.
		if m.Title == "" {
			m.Title = e.Title
		} else if preferTitle(m.Title, e.Title) {
			if !altSet[m.Title] {
				m.AltTitles = append(m.AltTitles, m.Title)
				altSet[m.Title] = true
			}
			m.Title = e.Title
		} else if e.Title != "" && e.Title != m.Title && !altSet[e.Title] {
			altSet[e.Title] = true
			m.AltTitles = append(m.AltTitles, e.Title)
		}
		for _, alt := range e.AltTitles {
			if alt == "" || alt == m.Title || altSet[alt] {
				continue
			}
			altSet[alt] = true
			m.AltTitles = append(m.AltTitles, alt)
		}

		if len(e.Description) > len(m.Description) {
			m.Description = e.Description
		}
		if e.Status != "" {
			m.Status = e.Status
		}
		for _, g := range e.Genres {
			if !genreSet[g] {
				genreSet[g] = true
				m.Genres = append(m.Genres, g)
			}
		}
		for _, t := range e.Tags {
			if !tagSet[t] {
				tagSet[t] = true
				m.Tags = append(m.Tags, t)
			}
		}
		if e.Chapters > m.Chapters {
			m.Chapters = e.Chapters
		}
		if e.Rating > 0 {
			ratingSum += e.Rating
			ratingN++
		}
		if e.Popularity > m.Popularity {
			m.Popularity = e.Popularity
		}
		if m.CoverURL == "" && e.CoverURL != "" {
			m.CoverURL = e.CoverURL
		}
		if e.UpdatedAt.After(m.UpdatedAt) {
			m.UpdatedAt = e.UpdatedAt
		}
	}

	if ratingN > 0 {
		m.Rating = clampRating(ratingSum / float64(ratingN))
	} else {
		m.Rating = 1 // model validation requires 1–10
	}
	if len(m.Genres) == 0 {
		m.Genres = []string{"unknown"}
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now().UTC()
	}
	return m
}

func clampRating(r float64) float64 {
	switch {
	case r < 1:
		return 1
	case r > 10:
		return 10
	default:
		return r
	}
}

// preferTitle reports whether `candidate` should replace `current` as the
// primary title — i.e. candidate is ASCII (Latin) but current is not.
func preferTitle(current, candidate string) bool {
	return isASCII(candidate) && !isASCII(current)
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}

// titleSimilar compares two manga titles loosely (lowercased, alphanumerics
// only) to decide whether two source rows describe the same work.
func titleSimilar(a, b string) bool {
	na := normalizeTitle(a)
	nb := normalizeTitle(b)
	if na == "" || nb == "" {
		return false
	}
	if na == nb {
		return true
	}
	if strings.HasPrefix(na, nb) || strings.HasPrefix(nb, na) {
		return true
	}
	return false
}

func normalizeTitle(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
