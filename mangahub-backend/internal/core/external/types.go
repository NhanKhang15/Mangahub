package external

import "time"

// SourceEntity is the normalized intermediate produced by every external
// client (MangaDex, MyAnimeList, AniList) before the Aggregator merges them
// into a canonical mangaModel.Manga. Fields that a particular source does not
// provide are simply left at their zero value.
type SourceEntity struct {
	Source      string    // "mangadex" | "myanimelist" | "anilist"
	ExternalID  string
	Title       string
	AltTitles   []string
	Description string
	Status      string // ongoing | completed | hiatus
	Genres      []string
	Tags        []string
	Authors     []string
	Artists     []string
	Chapters    int
	Rating      float64 // normalized 1-10
	Popularity  int     // higher = more popular
	CoverURL    string
	UpdatedAt   time.Time
}

// Aliases keep the public API aligned with the Phase 3 plan, which calls out
// MangaDexEntity / MALEntity / AniListEntity by name.
type (
	MangaDexEntity = SourceEntity
	MALEntity      = SourceEntity
	AniListEntity  = SourceEntity
)

// Chapter is a lightweight chapter record returned by MangaDex.
type Chapter struct {
	ID        string
	MangaID   string
	Number    string
	Title     string
	Language  string
	PublishAt time.Time
}
