# MangaHub Backend — API & Class Reference

> Phạm vi: REST API hiện có (Phase 0 + Phase 1). Cập nhật sau mỗi phase.
>
> - Base URL: `http://localhost:8080`
> - Content-Type: `application/json` cho mọi body
> - Auth tạm thời (Phase 1): header `X-User-ID: <hex objectid>` cho `/me/*`. Phase 2 sẽ thay bằng `Authorization: Bearer <jwt>`.
> - Mongo collection: `manga`, `artists`, `users`, `reading_progress`, `subscriptions`, `notifications` — xem [BACKEND_PLAN.md §3](../BACKEND_PLAN.md).

---

## Mục lục

1. [Error response chuẩn](#1-error-response-chuẩn)
2. [Constants / Enum](#2-constants--enum)
3. [Domain types (persistence model)](#3-domain-types-persistence-model)
4. [Request / Response DTOs (handler types)](#4-request--response-dtos-handler-types)
5. [REST endpoints](#5-rest-endpoints)
   - [Health](#51-health)
   - [Manga](#52-manga)
   - [Artists](#53-artists)
   - [Stats](#54-stats)
   - [Me — Reading progress](#55-me--reading-progress)
6. [Repository interfaces](#6-repository-interfaces)
7. [Service types](#7-service-types)
8. [File map](#8-file-map)

---

## 1. Error response chuẩn

Mọi response lỗi (4xx / 5xx) đều có shape:

```json
{
  "error": "<thông điệp con người đọc được>",
  "code":  "<MÃ_LỖI_VIẾT_HOA>",
  "details": { "field": "value", "...": "..." }
}
```

Định nghĩa trong Go: [internal/gateway/handler/error.go](internal/gateway/handler/error.go)

```go
type ErrorBody struct {
    Error   string         `json:"error"`
    Code    string         `json:"code"`
    Details map[string]any `json:"details,omitempty"`
}
```

| Code | HTTP | Khi nào |
|---|---|---|
| `INVALID_BODY` | 400 | JSON sai format hoặc validator fail (rating > 10, status sai enum, …) |
| `INVALID_QUERY` | 400 | Query string sai (page < 1, limit > 100, …) |
| `INVALID_ID` | 400 | Path param không phải ObjectID hex hợp lệ |
| `NO_FIELDS` | 400 | Body update rỗng (không có field nào để `$set`) |
| `UNAUTHORIZED` | 401 | Thiếu/sai `X-User-ID` (sẽ là JWT ở Phase 2) |
| `NOT_FOUND` | 404 | Mongo trả `ErrNoDocuments` |
| `INTERNAL` | 500 | Lỗi không xác định, kèm `details.cause` |

---

## 2. Constants / Enum

Định nghĩa trong package `internal/domain`:

```go
// Manga.Status — internal/domain/manga.go
const (
    MangaStatusOngoing   = "ongoing"
    MangaStatusCompleted = "completed"
    MangaStatusHiatus    = "hiatus"
)

// Artist.Role — internal/domain/artist.go
const (
    ArtistRoleArtist = "artist"
    ArtistRoleAuthor = "author"
    ArtistRoleBoth   = "both"
)

// ReadingProgress.Status — internal/domain/progress.go
const (
    ProgressReading    = "reading"
    ProgressCompleted  = "completed"
    ProgressPlanToRead = "plan_to_read"
    ProgressDropped    = "dropped"
)
```

| Trường | Giá trị hợp lệ |
|---|---|
| `Manga.status` | `ongoing`, `completed`, `hiatus` |
| `Manga.rating` | `1.0 ≤ x ≤ 10.0` |
| `Manga.chapters`, `popularity` | `≥ 0` |
| `Artist.role` | `artist`, `author`, `both` |
| `ReadingProgress.status` | `reading`, `completed`, `plan_to_read`, `dropped` |
| `ReadingProgress.rating` | tùy chọn, nếu có thì `1.0 ≤ x ≤ 10.0` |
| Pagination `page` | `≥ 1` |
| Pagination `limit` | `1 ≤ x ≤ 100` (mặc định `20`) |
| Stats `limit` | `1 ≤ x ≤ 100` (mặc định `10`) |

Sort key cho `GET /manga?sort=`:

| Giá trị | Hành vi |
|---|---|
| `popularity_desc` | `popularity` giảm dần |
| `rating_desc` | `rating` giảm dần |
| `title_asc` | `title` A→Z |
| `updated_at_desc` (default) | `updated_at` mới nhất trước |

---

## 3. Domain types (persistence model)

Đây là struct map 1-1 với document MongoDB. Cũng được serialize thẳng ra JSON cho client.

### `domain.Manga`
File: [internal/domain/manga.go](internal/domain/manga.go) — collection `manga`

```go
type Manga struct {
    ID          primitive.ObjectID   `json:"id"                     bson:"_id,omitempty"`
    ExternalIDs map[string]string    `json:"external_ids,omitempty" bson:"external_ids,omitempty"`
    Title       string               `json:"title"                  bson:"title"`
    AltTitles   []string             `json:"alt_titles,omitempty"   bson:"alt_titles,omitempty"`
    ArtistIDs   []primitive.ObjectID `json:"artist_ids,omitempty"   bson:"artist_ids,omitempty"`
    AuthorIDs   []primitive.ObjectID `json:"author_ids,omitempty"   bson:"author_ids,omitempty"`
    Description string               `json:"description,omitempty"  bson:"description,omitempty"`
    Status      string               `json:"status"                 bson:"status"`
    Genres      []string             `json:"genres"                 bson:"genres"`
    Tags        []string             `json:"tags,omitempty"         bson:"tags,omitempty"`
    Chapters    int                  `json:"chapters"               bson:"chapters"`
    Rating      float64              `json:"rating"                 bson:"rating"`
    CoverURL    string               `json:"cover_url,omitempty"    bson:"cover_url,omitempty"`
    Popularity  int                  `json:"popularity"             bson:"popularity"`
    CreatedAt   time.Time            `json:"created_at"             bson:"created_at"`
    UpdatedAt   time.Time            `json:"updated_at"             bson:"updated_at"`
}
```

JSON ví dụ:
```json
{
  "id": "69f03023891ae5d40fa99c20",
  "title": "One Piece",
  "external_ids": { "mangadex": "32d76d19-8a05" },
  "artist_ids": ["69f03034891ae5d40fa99c21"],
  "status": "ongoing",
  "genres": ["action","adventure"],
  "tags": ["shounen","pirates"],
  "chapters": 1092,
  "rating": 9.5,
  "popularity": 100,
  "created_at": "2026-04-28T03:57:23.282Z",
  "updated_at": "2026-04-28T03:57:23.282Z"
}
```

### `domain.MangaListQuery`
Tham số nội bộ cho repo `List`. Handler bind từ query string rồi convert.

```go
type MangaListQuery struct {
    Page  int
    Limit int
    Genre string
    Tags  []string
    Q     string
    Sort  string
}
```

### `domain.Artist`
File: [internal/domain/artist.go](internal/domain/artist.go) — collection `artists`

```go
type Artist struct {
    ID          primitive.ObjectID   `json:"id"                     bson:"_id,omitempty"`
    ExternalIDs map[string]string    `json:"external_ids,omitempty" bson:"external_ids,omitempty"`
    Name        string               `json:"name"                   bson:"name"`
    Role        string               `json:"role"                   bson:"role"`
    Bio         string               `json:"bio,omitempty"          bson:"bio,omitempty"`
    MangaIDs    []primitive.ObjectID `json:"manga_ids,omitempty"    bson:"manga_ids,omitempty"`
    CreatedAt   time.Time            `json:"created_at"             bson:"created_at"`
    UpdatedAt   time.Time            `json:"updated_at"             bson:"updated_at"`
}
```

### `domain.ReadingProgress`
File: [internal/domain/progress.go](internal/domain/progress.go) — collection `reading_progress`

```go
type ReadingProgress struct {
    ID             primitive.ObjectID `json:"id"               bson:"_id,omitempty"`
    UserID         primitive.ObjectID `json:"user_id"          bson:"user_id"`
    MangaID        primitive.ObjectID `json:"manga_id"         bson:"manga_id"`
    Status         string             `json:"status"           bson:"status"`
    CurrentChapter int                `json:"current_chapter"  bson:"current_chapter"`
    Rating         float64            `json:"rating,omitempty" bson:"rating,omitempty"`
    LastReadAt     time.Time          `json:"last_read_at"     bson:"last_read_at"`
}
```

Compound unique index `(user_id, manga_id)` đảm bảo mỗi cặp user × manga chỉ tồn tại một row.

---

## 4. Request / Response DTOs (handler types)

### Manga

File: [internal/gateway/handler/manga.go](internal/gateway/handler/manga.go)

```go
type CreateMangaInput struct {
    Title       string   `json:"title"        binding:"required,min=1,max=300"`
    AltTitles   []string `json:"alt_titles"`
    ArtistIDs   []string `json:"artist_ids"`        // hex ObjectID strings
    AuthorIDs   []string `json:"author_ids"`
    Description string   `json:"description"`
    Status      string   `json:"status"       binding:"required,oneof=ongoing completed hiatus"`
    Genres      []string `json:"genres"       binding:"required,min=1,dive,min=1"`
    Tags        []string `json:"tags"`
    Chapters    int      `json:"chapters"     binding:"gte=0"`
    Rating      float64  `json:"rating"       binding:"gte=1,lte=10"`
    CoverURL    string   `json:"cover_url"`
    Popularity  int      `json:"popularity"   binding:"gte=0"`
}

// Partial update — pointer = "field optional".
// Field nil ⇒ giữ nguyên. Field có giá trị ⇒ $set vào Mongo.
type UpdateMangaInput struct {
    Title       *string   `json:"title"        binding:"omitempty,min=1,max=300"`
    AltTitles   *[]string `json:"alt_titles"`
    ArtistIDs   *[]string `json:"artist_ids"`
    AuthorIDs   *[]string `json:"author_ids"`
    Description *string   `json:"description"`
    Status      *string   `json:"status"       binding:"omitempty,oneof=ongoing completed hiatus"`
    Genres      *[]string `json:"genres"       binding:"omitempty,min=1"`
    Tags        *[]string `json:"tags"`
    Chapters    *int      `json:"chapters"     binding:"omitempty,gte=0"`
    Rating      *float64  `json:"rating"       binding:"omitempty,gte=1,lte=10"`
    CoverURL    *string   `json:"cover_url"`
    Popularity  *int      `json:"popularity"   binding:"omitempty,gte=0"`
}

type ListMangaQuery struct {
    Page  int    `form:"page,default=1"   binding:"gte=1"`
    Limit int    `form:"limit,default=20" binding:"gte=1,lte=100"`
    Genre string `form:"genre"`
    Tags  string `form:"tags"`   // CSV — handler tự split bằng dấu phẩy
    Q     string `form:"q"`
    Sort  string `form:"sort"`
}
```

### Artist

File: [internal/gateway/handler/artist.go](internal/gateway/handler/artist.go)

```go
type CreateArtistInput struct {
    Name string `json:"name" binding:"required,min=1,max=200"`
    Role string `json:"role" binding:"required,oneof=artist author both"`
    Bio  string `json:"bio"`
}

type UpdateArtistInput struct {
    Name *string `json:"name" binding:"omitempty,min=1,max=200"`
    Role *string `json:"role" binding:"omitempty,oneof=artist author both"`
    Bio  *string `json:"bio"`
}

type ListArtistQuery struct {
    Page  int    `form:"page,default=1"   binding:"gte=1"`
    Limit int    `form:"limit,default=20" binding:"gte=1,lte=100"`
    Q     string `form:"q"`
}

type ListArtistMangaQuery struct {
    Page  int `form:"page,default=1"   binding:"gte=1"`
    Limit int `form:"limit,default=20" binding:"gte=1,lte=100"`
}
```

### Stats

File: [internal/gateway/handler/stats.go](internal/gateway/handler/stats.go)

```go
type LimitQuery struct {
    Limit int `form:"limit,default=10" binding:"gte=1,lte=100"`
}
```

### Reading progress

File: [internal/gateway/handler/progress.go](internal/gateway/handler/progress.go)

```go
type UpsertProgressInput struct {
    Status         string  `json:"status"          binding:"required,oneof=reading completed plan_to_read dropped"`
    CurrentChapter int     `json:"current_chapter" binding:"gte=0"`
    Rating         float64 `json:"rating"          binding:"omitempty,gte=1,lte=10"`
}

type ListProgressQuery struct {
    Page   int    `form:"page,default=1"   binding:"gte=1"`
    Limit  int    `form:"limit,default=20" binding:"gte=1,lte=100"`
    Status string `form:"status"           binding:"omitempty,oneof=reading completed plan_to_read dropped"`
}
```

### Pagination response shape

Tất cả endpoint list trả về cùng shape:

```json
{
  "data":  [ /* []Entity */ ],
  "page":  1,
  "limit": 20,
  "total": 342
}
```

Endpoint stats / `/me/stats` trả `{"data": ...}` không có pagination.

---

## 5. REST endpoints

### 5.1 Health

#### `GET /healthz`
Ping Mongo. Không cần auth.

**Response 200**
```json
{ "status": "ok", "deps": { "mongo": "ok" } }
```

**Response 503** (Mongo down)
```json
{ "status": "degraded", "deps": { "mongo": "context deadline exceeded" } }
```

---

### 5.2 Manga

Base path: `/manga` — tất cả public (Phase 1, chưa có role admin).

#### `GET /manga`
List manga với filter + pagination + search.

**Query params** ⇒ [ListMangaQuery](#manga)

| Param | Default | Ghi chú |
|---|---|---|
| `page` | 1 | `≥ 1` |
| `limit` | 20 | `1 ≤ x ≤ 100` |
| `genre` | — | filter exact (1 genre) |
| `tags` | — | CSV, ví dụ `tags=shounen,pirates` (yêu cầu document chứa **tất cả** tags) |
| `q` | — | text search (dùng text index trên `title` + `description`) |
| `sort` | `updated_at_desc` | xem [§2](#2-constants--enum) |

**Response 200** — pagination shape, `data: []Manga`.

**Lỗi:** `400 INVALID_QUERY` nếu `page < 1` hoặc `limit > 100`.

#### `GET /manga/:id`
Lấy 1 manga.

| Path param | Loại |
|---|---|
| `id` | hex ObjectID |

**Response 200** — `Manga`.
**Response 400 INVALID_ID** — id không phải hex.
**Response 404 NOT_FOUND**.

#### `POST /manga`
Tạo manga.

**Body** ⇒ [CreateMangaInput](#manga)
```json
{
  "title": "One Piece",
  "status": "ongoing",
  "genres": ["action","adventure"],
  "chapters": 1092,
  "rating": 9.5,
  "popularity": 100,
  "artist_ids": ["69f03034891ae5d40fa99c21"]
}
```

**Response 201** — `Manga` đầy đủ với `id`, `created_at`, `updated_at`.
**Response 400** — `INVALID_BODY` nếu validator fail; `INVALID_ID` nếu phần tử trong `artist_ids`/`author_ids` không phải hex.

#### `PUT /manga/:id`
Partial update — chỉ `$set` các field có trong body.

**Body** ⇒ [UpdateMangaInput](#manga). Các field đều optional, nhưng tổng phải có ít nhất 1 field.

**Response 200** — `Manga` sau update.
**Response 400** — `INVALID_BODY`, `INVALID_ID`, hoặc `NO_FIELDS`.
**Response 404 NOT_FOUND**.

#### `DELETE /manga/:id`
**Response 204** (no content).
**Response 400 INVALID_ID** / **404 NOT_FOUND**.

---

### 5.3 Artists

Base path: `/artists`.

#### `GET /artists`
**Query** ⇒ [ListArtistQuery](#artist). `q` dùng text index trên `name`.
**Response 200** — pagination shape, `data: []Artist`.

#### `GET /artists/:id`
**Response 200** — `Artist`. **404** nếu không tồn tại.

#### `POST /artists`
**Body** ⇒ [CreateArtistInput](#artist)
```json
{ "name": "Eiichiro Oda", "role": "both", "bio": "Author of One Piece" }
```
**Response 201** — `Artist`.

#### `PUT /artists/:id`
**Body** ⇒ [UpdateArtistInput](#artist). Pattern partial update giống Manga.
**Response 200**.

#### `DELETE /artists/:id`
**Response 204**.

#### `GET /artists/:id/manga`
List manga có `artist_ids` chứa `:id`. Đây là cách MangaHub mô phỏng quan hệ Artist→Manga theo lab Books.

**Query** ⇒ [ListArtistMangaQuery](#artist). Sort cố định `updated_at desc`.

**Response 200**
```json
{
  "data": [ /* []Manga */ ],
  "page": 1,
  "limit": 20,
  "total": 7
}
```

---

### 5.4 Stats

Base path: `/stats`.

#### `GET /stats/popular`
Top manga theo `popularity` desc.

**Query** ⇒ [LimitQuery](#stats) (`limit` default 10, max 100).

**Response 200**
```json
{ "data": [ /* []Manga */ ] }
```

#### `GET /stats/trending`
Manga có `updated_at` trong 30 ngày gần nhất, sort theo `popularity` desc rồi `updated_at` desc.

Cùng query + response shape với `/stats/popular`.

---

### 5.5 Me — Reading progress

Base path: `/me`. **Tất cả endpoint trong nhóm này yêu cầu header**:

```
X-User-ID: 65f0a000a000a000a000a000
```

(Phase 2 sẽ thay bằng `Authorization: Bearer <jwt>` — chỉ phải đổi [middleware/user.go](internal/gateway/middleware/user.go).)

Thiếu/sai header ⇒ `401 UNAUTHORIZED`.

#### `GET /me/reading`
**Query** ⇒ [ListProgressQuery](#reading-progress). Filter optional theo `status`. Sort cố định `last_read_at desc`.

**Response 200** — pagination shape, `data: []ReadingProgress`.

#### `PUT /me/reading/:mangaId`
Upsert progress cho user hiện tại × manga `:mangaId`. Tự set `last_read_at = now()`.

**Body** ⇒ [UpsertProgressInput](#reading-progress)
```json
{ "status": "reading", "current_chapter": 500, "rating": 10 }
```

**Response 200** — `ReadingProgress` sau upsert.

#### `DELETE /me/reading/:mangaId`
Xóa entry. **Response 204**. **404** nếu chưa có.

#### `GET /me/stats`
Đếm số progress theo từng status (aggregation `$group`).

**Response 200**
```json
{ "data": { "reading": 12, "completed": 3, "plan_to_read": 5 } }
```

Field nào không có document sẽ không xuất hiện trong response.

---

## 6. Repository interfaces

Nguồn của sự thật cho data access. Mỗi service inject 1 implementation (hiện tại chỉ có `mongoRepo`); Phase 4 sẽ giữ interface, đổi implementation thành gRPC client mà handler không phải thay.

### `catalog.Repo`
File: [internal/catalog/repo.go](internal/catalog/repo.go)

```go
type Repo interface {
    Create(ctx context.Context, m *domain.Manga) (primitive.ObjectID, error)
    Get(ctx context.Context, id primitive.ObjectID) (*domain.Manga, error)
    Update(ctx context.Context, id primitive.ObjectID, set bson.M) (*domain.Manga, error)
    Delete(ctx context.Context, id primitive.ObjectID) error
    List(ctx context.Context, q domain.MangaListQuery) ([]*domain.Manga, int64, error)
    ListByArtist(ctx context.Context, artistID primitive.ObjectID, page, limit int) ([]*domain.Manga, int64, error)
    Popular(ctx context.Context, limit int) ([]*domain.Manga, error)
    Trending(ctx context.Context, limit int) ([]*domain.Manga, error)
}

var ErrNotFound = errors.New("manga not found")

func NewMongoRepo(db *mongo.Database) Repo
```

### `artist.Repo`
File: [internal/artist/repo.go](internal/artist/repo.go)

```go
type Repo interface {
    Create(ctx context.Context, a *domain.Artist) (primitive.ObjectID, error)
    Get(ctx context.Context, id primitive.ObjectID) (*domain.Artist, error)
    Update(ctx context.Context, id primitive.ObjectID, set bson.M) (*domain.Artist, error)
    Delete(ctx context.Context, id primitive.ObjectID) error
    List(ctx context.Context, q string, page, limit int) ([]*domain.Artist, int64, error)
}

var ErrNotFound = errors.New("artist not found")

func NewMongoRepo(db *mongo.Database) Repo
```

### `progress.Repo`
File: [internal/progress/repo.go](internal/progress/repo.go)

```go
type Repo interface {
    Upsert(ctx context.Context, p *domain.ReadingProgress) (*domain.ReadingProgress, error)
    Get(ctx context.Context, userID, mangaID primitive.ObjectID) (*domain.ReadingProgress, error)
    List(ctx context.Context, userID primitive.ObjectID, status string, page, limit int) ([]*domain.ReadingProgress, int64, error)
    Delete(ctx context.Context, userID, mangaID primitive.ObjectID) error
    Stats(ctx context.Context, userID primitive.ObjectID) (map[string]int, error)
}

var ErrNotFound = errors.New("reading progress not found")

func NewMongoRepo(db *mongo.Database) Repo
```

---

## 7. Service types

Service mỏng — chủ yếu pass-through xuống repo. Tồn tại để (a) gắn business rules sau này (vd: tăng `popularity` mỗi khi có người đọc), (b) tách handler khỏi repo trong test.

```go
// internal/catalog/service.go
type Service struct{ /* repo Repo */ }
func NewService(r Repo) *Service

// internal/artist/service.go
type Service struct{ /* repo Repo */ }
func NewService(r Repo) *Service

// internal/progress/service.go
type Service struct{ /* repo Repo */ }
func NewService(r Repo) *Service
```

Wiring trong [cmd/gateway/main.go](cmd/gateway/main.go):

```go
mangaSvc    := catalog.NewService(catalog.NewMongoRepo(cl.DB))
artistSvc   := artist.NewService(artist.NewMongoRepo(cl.DB))
progressSvc := progress.NewService(progress.NewMongoRepo(cl.DB))

r := gateway.NewRouter(cfg.Env, gateway.Deps{
    MongoClient: cl.Mongo,
    MangaSvc:    mangaSvc,
    ArtistSvc:   artistSvc,
    ProgressSvc: progressSvc,
})
```

`gateway.Deps` (file [internal/gateway/router.go](internal/gateway/router.go)):

```go
type Deps struct {
    MongoClient *mongo.Client
    MangaSvc    *catalog.Service
    ArtistSvc   *artist.Service
    ProgressSvc *progress.Service
}
```

---

## 8. File map

| Trách nhiệm | File |
|---|---|
| Entry point + DI | [cmd/gateway/main.go](cmd/gateway/main.go) |
| Mount routes | [internal/gateway/router.go](internal/gateway/router.go) |
| Mongo connect + indexes | [internal/platform/mongo/mongo.go](internal/platform/mongo/mongo.go) |
| Config loader (.env + env) | [internal/platform/config/config.go](internal/platform/config/config.go) |
| CORS | [internal/gateway/middleware/cors.go](internal/gateway/middleware/cors.go) |
| Auth header (placeholder) | [internal/gateway/middleware/user.go](internal/gateway/middleware/user.go) |
| Error helper + ObjectID parser | [internal/gateway/handler/error.go](internal/gateway/handler/error.go) |
| `/healthz` | [internal/gateway/handler/health.go](internal/gateway/handler/health.go) |
| `/manga` handlers | [internal/gateway/handler/manga.go](internal/gateway/handler/manga.go) |
| `/artists` handlers | [internal/gateway/handler/artist.go](internal/gateway/handler/artist.go) |
| `/stats` handlers | [internal/gateway/handler/stats.go](internal/gateway/handler/stats.go) |
| `/me/reading` handlers | [internal/gateway/handler/progress.go](internal/gateway/handler/progress.go) |
| Manga domain + persistence model | [internal/domain/manga.go](internal/domain/manga.go) |
| Artist domain | [internal/domain/artist.go](internal/domain/artist.go) |
| Progress domain | [internal/domain/progress.go](internal/domain/progress.go) |
| Manga repo (Mongo) | [internal/catalog/repo.go](internal/catalog/repo.go) |
| Artist repo | [internal/artist/repo.go](internal/artist/repo.go) |
| Progress repo | [internal/progress/repo.go](internal/progress/repo.go) |
| Manga service | [internal/catalog/service.go](internal/catalog/service.go) |
| Artist service | [internal/artist/service.go](internal/artist/service.go) |
| Progress service | [internal/progress/service.go](internal/progress/service.go) |

---

## Phụ lục — curl cheatsheet

```bash
BASE=http://localhost:8080
UID=65f0a000a000a000a000a000

# health
curl -s $BASE/healthz

# create manga
MID=$(curl -s -X POST $BASE/manga -H "Content-Type: application/json" \
  -d '{"title":"One Piece","status":"ongoing","genres":["action"],"chapters":1092,"rating":9.5,"popularity":100}' \
  | jq -r .id)

# list + filter + search
curl -s "$BASE/manga?genre=action&page=1&limit=10&sort=popularity_desc"
curl -s "$BASE/manga?q=piece"

# update partial
curl -s -X PUT $BASE/manga/$MID -H "Content-Type: application/json" \
  -d '{"chapters":1093,"popularity":110}'

# stats
curl -s "$BASE/stats/popular?limit=5"
curl -s "$BASE/stats/trending?limit=5"

# create artist + relate
AID=$(curl -s -X POST $BASE/artists -H "Content-Type: application/json" \
  -d '{"name":"Eiichiro Oda","role":"both"}' | jq -r .id)
curl -s -X PUT $BASE/manga/$MID -H "Content-Type: application/json" \
  -d "{\"artist_ids\":[\"$AID\"]}"
curl -s "$BASE/artists/$AID/manga"

# reading progress
curl -s -X PUT "$BASE/me/reading/$MID" \
  -H "Content-Type: application/json" -H "X-User-ID: $UID" \
  -d '{"status":"reading","current_chapter":500,"rating":10}'
curl -s -H "X-User-ID: $UID" "$BASE/me/reading?status=reading"
curl -s -H "X-User-ID: $UID" "$BASE/me/stats"
```
