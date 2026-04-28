# MangaHub — Backend Plan (Go)

> Phạm vi: chỉ backend. Frontend sẽ làm bằng ngôn ngữ khác và gọi vào REST + WebSocket của service này. gRPC là internal (service-to-service).

---

## 1. Tech stack & quyết định kiến trúc

| Thành phần | Lựa chọn | Lý do |
|---|---|---|
| Ngôn ngữ | Go 1.22+ | Toàn bộ lab đều dùng Go |
| HTTP framework | `gin-gonic/gin` (hoặc `chi`) | Routing + middleware gọn, đủ cho CRUD |
| WebSocket | `gorilla/websocket` | Đúng chuẩn lab Chat |
| gRPC | `google.golang.org/grpc` + `protoc-gen-go` | Đúng chuẩn lab gRPC |
| DB | MongoDB 7 (driver `go.mongodb.org/mongo-driver`) | Đúng yêu cầu (NoSQL lab) |
| Auth | JWT (`golang-jwt/jwt/v5`) | Đề bài bắt buộc dùng JWT cho WS |
| Config | `viper` hoặc `env` thuần | Đọc `.env` cho secrets |
| Logging | `log/slog` (chuẩn lib) | Có structured log |
| Validate | `go-playground/validator/v10` | Validate rating 1-10, query params |
| HTTP client | `net/http` + `golang.org/x/time/rate` | Rate limit khi gọi API ngoài |
| Container | Docker + docker-compose | Chạy MongoDB + 3 services |

**Quyết định kiến trúc:**
- 1 binary REST gateway (HTTP + WebSocket) + N binary gRPC services. REST gateway **không** truy cập MongoDB trực tiếp ngoài hub WS — mọi nghiệp vụ data đi qua gRPC.
- gRPC services tự kết nối MongoDB (mỗi service một collection chính của nó).
- WebSocket Hub chạy chung process với REST gateway (cùng cần JWT middleware).

```
                ┌──────────────────────────────────────┐
   Frontend ──► │  api-gateway (HTTP REST + WebSocket) │
                │  :8080  /manga, /ws, /auth …         │
                └──────┬─────────────┬─────────────────┘
                       │ gRPC        │ gRPC
                       ▼             ▼
              ┌────────────┐  ┌────────────────┐
              │ catalog-svc│  │ artist-svc     │
              │  :50051    │  │  :50052        │
              └─────┬──────┘  └──────┬─────────┘
                    │                │
                    ▼                ▼
              ┌────────────┐  ┌────────────────┐
              │ progress-  │  │ user-prefs-svc │
              │ svc :50053 │  │  :50054        │
              └─────┬──────┘  └──────┬─────────┘
                    └────────┬───────┘
                             ▼
                       MongoDB :27017
```

---

## 2. Cấu trúc thư mục đề xuất

```
mangahub-backend/
├── cmd/
│   ├── gateway/main.go            # REST + WebSocket
│   ├── catalog-svc/main.go        # gRPC manga catalog
│   ├── artist-svc/main.go         # gRPC artist/author
│   ├── progress-svc/main.go       # gRPC reading progress
│   └── prefs-svc/main.go          # gRPC user preferences
├── internal/
│   ├── gateway/
│   │   ├── handler/               # gin handlers (manga, artist, auth, stats)
│   │   ├── middleware/            # jwt, ratelimit, cors, logger
│   │   ├── ws/                    # hub.go, client.go, room.go (lab Chat)
│   │   └── grpcclient/            # dial wrappers tới các service
│   ├── catalog/                   # business logic + mongo repo cho manga
│   ├── artist/
│   ├── progress/
│   ├── prefs/
│   ├── external/                  # MangaDex / MAL / AniList clients
│   │   ├── mangadex.go
│   │   ├── myanimelist.go
│   │   ├── anilist.go
│   │   └── aggregator.go
│   ├── domain/                    # entity thuần (Manga, Artist, Chapter…)
│   ├── auth/                      # JWT issue/verify
│   └── platform/
│       ├── mongo/                 # connect, indexes
│       └── config/
├── proto/                         # .proto files + generated *.pb.go
│   ├── catalog.proto
│   ├── artist.proto
│   ├── progress.proto
│   └── prefs.proto
├── deploy/
│   ├── docker-compose.yml
│   └── Dockerfile.{gateway,catalog,artist,progress,prefs}
├── scripts/
│   ├── seed.go                    # seed manga từ MangaDex vào Mongo
│   └── proto-gen.sh
├── test/                          # integration tests
├── .env.example
├── go.mod
└── Makefile
```

---

## 3. Domain model & MongoDB schema

### Collections

**`manga`**
```go
type Manga struct {
    ID          primitive.ObjectID `bson:"_id"`
    ExternalIDs map[string]string  `bson:"external_ids"`   // {"mangadex":"...","mal":"...","anilist":"..."}
    Title       string             `bson:"title"`
    AltTitles   []string           `bson:"alt_titles"`
    ArtistIDs   []primitive.ObjectID `bson:"artist_ids"`
    AuthorIDs   []primitive.ObjectID `bson:"author_ids"`
    Description string             `bson:"description"`
    Status      string             `bson:"status"`          // ongoing|completed|hiatus
    Genres      []string           `bson:"genres"`
    Tags        []string           `bson:"tags"`
    Chapters    int                `bson:"chapters"`
    Rating      float64            `bson:"rating"`          // 1-10, validate
    CoverURL    string             `bson:"cover_url"`
    Popularity  int                `bson:"popularity"`      // dùng cho trending
    CreatedAt   time.Time          `bson:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at"`
}
```
Indexes: `{title:"text", description:"text"}`, `{genres:1}`, `{tags:1}`, `{popularity:-1}`, `{external_ids.mangadex:1}` (unique sparse).

**`artists`** — `{_id, name, role:"artist|author|both", external_ids, bio, manga_ids[]}`. Index `{name:"text"}`, `{role:1}`.

**`users`** — `{_id, username, email, password_hash, created_at}`. Index `{email:1}` unique, `{username:1}` unique.

**`reading_progress`** — `{_id, user_id, manga_id, status:"reading|completed|plan_to_read|dropped", current_chapter, last_read_at, rating?}`. Index `{user_id:1, manga_id:1}` unique compound.

**`subscriptions`** — `{_id, user_id, room, created_at}` với `room` = `manga:<id>` hoặc `genre:<name>`. Index `{user_id:1}`, `{room:1}`.

**`notifications`** (optional, để replay khi user offline) — `{_id, user_id, type, payload, sent_at, delivered:bool}`.

---

## 4. Roadmap theo phase (8–9 tuần)

> Mỗi phase nên kết thúc bằng 1 demo chạy được + commit + 1 đoạn test cơ bản.

### Phase 0 — Bootstrap (2–3 ngày)
- [ ] `go mod init`, dựng `Makefile` (build/run/test/proto-gen).
- [ ] `docker-compose.yml` chạy MongoDB + Mongo Express.
- [ ] `internal/platform/mongo` connect + ping + tạo indexes lúc startup.
- [ ] `internal/platform/config` đọc `.env` (MONGO_URI, JWT_SECRET, MANGADEX_BASE, ports…).
- [ ] Health-check endpoint `GET /healthz`.

### Phase 1 — Domain + REST CRUD (lab Books) (1 tuần)
Mục tiêu: REST gateway nói chuyện trực tiếp với Mongo trước, **chưa tách gRPC**. Mục đích là có API chạy được sớm để test.
- [ ] Domain structs trong `internal/domain`.
- [ ] Repo `manga_repo.go`, `artist_repo.go` dùng mongo driver (CRUD + filter + paginate).
- [ ] Handlers theo đúng spec đề bài:
  - `GET /manga?page=&limit=&genre=&tags=&q=` (pagination + search + filter)
  - `GET /manga/:id`
  - `POST /manga`, `PUT /manga/:id`, `DELETE /manga/:id`
  - `GET /artists/:id/manga`
  - `GET /stats/popular`, `GET /stats/trending` (sort by `popularity`, by `updated_at` 30 ngày gần nhất)
- [ ] Validator: rating 1–10, status enum, pagination caps (`limit ≤ 100`).
- [ ] Reading status endpoints (cho đến khi tách gRPC sẽ chuyển):
  - `PUT /me/reading/:mangaId` body `{status, current_chapter, rating?}`
  - `GET /me/reading?status=`
- [ ] Middleware: logger, recover, CORS, error JSON chuẩn `{error, code, details}`.
- [ ] Unit test cho repo (testcontainers mongo) + handler test (httptest).

**Deliverable:** Postman collection chạy đủ CRUD + filter + pagination.

### Phase 2 — Auth JWT (2–3 ngày)
- [ ] `POST /auth/register`, `POST /auth/login` → trả access token (15m) + refresh token (7d).
- [ ] `internal/auth`: hash password bằng `bcrypt`, sign JWT HS256.
- [ ] Middleware `RequireAuth` đọc `Authorization: Bearer <token>`, gắn `userID` vào `ctx`.
- [ ] Đổi các route `/me/...` sang require auth.

### Phase 3 — Data Collection Layer (lab TMDB) (1 tuần)
- [ ] `internal/external/mangadex.go`:
  ```go
  type MangaDexClient struct {
      BaseURL    string
      HTTPClient *http.Client
      AuthToken  string
      limiter    *rate.Limiter   // 5 req/s mặc định
  }
  func (c *MangaDexClient) SearchManga(ctx, q string, page int) ([]MangaDexEntity, error)
  func (c *MangaDexClient) GetManga(ctx, id string) (*MangaDexEntity, error)
  func (c *MangaDexClient) GetChapters(ctx, id string) ([]Chapter, error)
  ```
  - Retry 3 lần với exponential backoff khi 429/5xx.
  - Parse `Retry-After` nếu có.
- [ ] `myanimelist.go`, `anilist.go` cùng pattern (AniList dùng GraphQL → 1 helper `Post` JSON).
- [ ] `aggregator.go`:
  - Gọi 3 nguồn song song bằng `errgroup`.
  - Merge theo title-similarity hoặc external IDs.
  - Mapping field: `Director→Author`, `Runtime→Chapters`, MAL `score`/AniList `averageScore` → `Rating` (chuẩn hóa thang 1–10).
- [ ] Endpoint admin `POST /admin/import?source=mangadex&q=one+piece` (yêu cầu role admin) → upsert vào Mongo theo `external_ids.<source>`.
- [ ] Cron seed (script `scripts/seed.go`) chạy 1 lần lấy top 100 trending.

**Deliverable:** Có thể `curl POST /admin/import` để bulk-load manga thật từ MangaDex.

### Phase 4 — Tách microservice & gRPC (lab gRPC) (1.5 tuần)
- [ ] Viết `proto/catalog.proto`:
  ```proto
  service MangaCatalog {
    rpc GetManga(GetMangaRequest) returns (Manga);
    rpc CreateManga(CreateMangaRequest) returns (Manga);
    rpc UpdateManga(UpdateMangaRequest) returns (Manga);
    rpc DeleteManga(DeleteMangaRequest) returns (google.protobuf.Empty);
    rpc ListManga(ListMangaRequest) returns (ListMangaResponse);  // có page_token + filter
    rpc SearchManga(SearchMangaRequest) returns (ListMangaResponse);
  }
  ```
- [ ] `artist.proto`: `GetArtist`, `ListArtistManga`, `CreateArtist`, `SearchArtist`.
- [ ] `progress.proto`: `UpsertProgress`, `GetProgress`, `ListUserProgress`, `Stats`.
- [ ] `prefs.proto`: `GetPreferences`, `UpdatePreferences`, `Subscribe`, `Unsubscribe`, `ListSubscriptions` (subscription cho WS rooms).
- [ ] Generate code: `scripts/proto-gen.sh` (xài `buf` nếu quen, hoặc `protoc` thuần).
- [ ] Implement server cho mỗi service trong `cmd/<svc>/main.go`, dùng repo trong `internal/<svc>`.
- [ ] gRPC interceptors: logging, recovery, auth (forward `x-user-id` từ gateway qua metadata).
- [ ] gRPC error → map sang HTTP status ở gateway (`codes.NotFound` → 404, `InvalidArgument` → 400…).
- [ ] Refactor handlers ở gateway: thay vì gọi repo Mongo, dial gRPC service. Pool connection per service, có `grpc.WithDefaultServiceConfig` retry policy.

**Deliverable:** `docker-compose up` lên 5 container (4 gRPC + 1 gateway + 1 mongo). Tất cả endpoint REST cũ vẫn pass test.

### Phase 5 — WebSocket real-time (lab Chat) (1 tuần)
- [ ] `internal/gateway/ws/hub.go` — Hub pattern lab Chat:
  ```go
  type Hub struct {
      register   chan *Client
      unregister chan *Client
      broadcast  chan *RoomMessage      // theo room
      direct     chan *DirectMessage    // tới user cụ thể
      rooms      map[string]map[*Client]bool
      byUser     map[string]*Client     // userID -> client
  }
  ```
- [ ] `client.go` — `readPump`/`writePump`, ping/pong 54s/60s, write deadline 10s.
- [ ] Endpoint `GET /ws` upgrade WebSocket. **Auth bằng JWT trong header `Sec-WebSocket-Protocol` hoặc query `?token=` rồi verify** — đề bài yêu cầu JWT, không phải query param raw username, nên bắt buộc verify JWT trước khi upgrade.
- [ ] Message protocol JSON:
  ```json
  // client → server
  {"type":"subscribe","room":"manga:one-piece"}
  {"type":"unsubscribe","room":"manga:one-piece"}
  {"type":"ping"}

  // server → client
  {"type":"new_chapter","manga":"one-piece","content":"Chapter 1092 released!","chapter":1092,"ts":"..."}
  {"type":"recommendation","manga":"naruto","content":"Friend X recommended"}
  {"type":"system","content":"connected"}
  ```
- [ ] `Subscribe` đồng bộ persist xuống `subscriptions` collection qua `prefs-svc` (gRPC) → reload được khi reconnect.
- [ ] Trigger `new_chapter`:
  - Job `internal/gateway/poller` mỗi 5 phút gọi MangaDex `GetChapters` cho từng manga đang có subscriber → diff với `chapters` trong DB → nếu mới thì:
    - Update DB qua `catalog-svc.UpdateManga`.
    - Push `Hub.broadcast <- RoomMessage{Room:"manga:<id>", ...}`.
- [ ] Direct notification: endpoint nội bộ `POST /admin/notify` (admin only) → `Hub.direct <- DirectMessage{UserID, ...}`.
- [ ] Reconnect strategy: server cấp `last_event_id`, client gửi lại khi reconnect → server replay từ `notifications` collection (đơn giản: chỉ replay 24h gần nhất).
- [ ] Test bằng `wscat` hoặc 1 file `cmd/wsclient/main.go` đơn giản.

**Deliverable:** 2 client cùng subscribe `manga:one-piece`, kích hoạt poller giả → cả 2 nhận được message.

### Phase 6 — Hoàn thiện features (3–4 ngày)
- [ ] **Statistics**: `GET /stats/popular` (top theo popularity), `GET /stats/trending` (manga có nhiều `reading_progress` update trong 7 ngày), `GET /me/stats` (số manga đang đọc, đã đọc, tổng chapter đọc).
- [ ] **Chapter tracking**: `PUT /me/reading/:mangaId/chapter` body `{chapter:int}`. Update qua `progress-svc`.
- [ ] **Recommendation đơn giản**: dựa trên genre của các manga user đã `completed` rating ≥ 8 → trả top manga cùng genre user chưa đọc. Endpoint `GET /me/recommendations`.
- [ ] **Rate limiting** trên gateway (per IP + per user) bằng `golang.org/x/time/rate`.
- [ ] **Admin role**: thêm field `roles []string` vào `users`, middleware `RequireRole("admin")`.

### Phase 7 — Test, Docker, docs (3–4 ngày)
- [ ] Integration test: chạy compose test, gọi end-to-end (REST → gRPC → Mongo, WS pub/sub).
- [ ] Unit test cho aggregator (mock HTTP với `httptest.Server`).
- [ ] `Dockerfile` multi-stage cho từng binary.
- [ ] `docker-compose.yml` final với healthcheck + depends_on.
- [ ] OpenAPI spec (`api/openapi.yaml`) — frontend bạn cần cái này.
- [ ] README chạy local + chạy docker + ví dụ curl.

---

## 5. Hợp đồng API REST (đầy đủ cho frontend)

| Method | Path | Auth | Mô tả |
|---|---|---|---|
| POST | `/auth/register` | – | `{username,email,password}` → tokens |
| POST | `/auth/login` | – | `{email,password}` → tokens |
| POST | `/auth/refresh` | refresh | đổi access token mới |
| GET | `/manga` | – | query: `page,limit,genre,tags,q,sort` |
| GET | `/manga/:id` | – | chi tiết, kèm `chapters_total`, `artists[]` |
| POST | `/manga` | admin | tạo manga |
| PUT | `/manga/:id` | admin | cập nhật |
| DELETE | `/manga/:id` | admin | xóa |
| GET | `/artists` | – | list + search |
| GET | `/artists/:id` | – | chi tiết |
| GET | `/artists/:id/manga` | – | manga của artist |
| GET | `/stats/popular` | – | top theo popularity |
| GET | `/stats/trending` | – | trending 7 ngày |
| GET | `/me/reading` | user | filter `?status=` |
| PUT | `/me/reading/:mangaId` | user | upsert reading status + chapter + rating |
| DELETE | `/me/reading/:mangaId` | user | bỏ khỏi list |
| GET | `/me/subscriptions` | user | list rooms đã subscribe |
| GET | `/me/recommendations` | user | manga gợi ý |
| POST | `/admin/import` | admin | import từ external source |
| POST | `/admin/notify` | admin | gửi direct notification |
| GET | `/ws` | user (JWT) | upgrade WebSocket |
| GET | `/healthz` | – | health |

**Response chuẩn:**
```json
// success list
{"data":[...], "page":1, "limit":20, "total":342}
// error
{"error":"manga not found", "code":"NOT_FOUND", "details":{}}
```

---

## 6. Mapping yêu cầu đề bài → phase/file

| Yêu cầu đề bài | Phase | File chính |
|---|---|---|
| MangaDex API Client (struct giống TMDBClient) | 3 | `internal/external/mangadex.go` |
| Multi-source aggregator (MangaDex+MAL+AniList) | 3 | `internal/external/aggregator.go` |
| Rate limiting external API | 3 | `internal/external/*` (`rate.Limiter`) |
| Mapping Director→Author, Runtime→Chapters | 3 | `aggregator.go` mapping fn |
| CRUD manga + artist | 1, 4 | `internal/catalog`, `internal/artist` |
| Search & filter genre/tags | 1, 4 | repo + handler |
| Pagination | 1, 4 | repo `Find().Skip().Limit()` + page_token gRPC |
| Reading status (đang đọc / đã đọc / sẽ đọc) | 1→4 | `internal/progress` |
| Chapter tracking | 4, 6 | `progress-svc` |
| Statistics popular/trending/reading | 6 | `internal/catalog`, `internal/progress` |
| Rating validation 1–10 | 1 | validator tag `gte=1,lte=10` |
| Hub pattern readPump/writePump | 5 | `internal/gateway/ws/hub.go`, `client.go` |
| Room-based broadcast (manga/genre) | 5 | `hub.broadcast` |
| Direct notification | 5 | `hub.direct` |
| Ping/pong keep-alive | 5 | `client.go` |
| Reconnection + replay | 5 | `notifications` collection |
| Auth WS bằng JWT | 5 | middleware trước Upgrade |
| Replace in-memory bằng Mongo | 5 | subscriptions persist qua `prefs-svc` |
| 4 gRPC service | 4 | `cmd/{catalog,artist,progress,prefs}-svc` |
| .proto files | 4 | `proto/*.proto` |
| MongoDB primary | 0+ | `internal/platform/mongo` |
| Collections: manga/artists/users/reading_progress/subscriptions | 0+ | schema mục 3 |

---

## 7. Lệnh chạy nhanh

```bash
# proto
make proto

# dev (compose mongo + tất cả service)
make up

# chạy 1 service riêng
go run ./cmd/gateway
go run ./cmd/catalog-svc

# test
make test           # unit
make test-int       # integration (cần docker)

# seed manga
go run ./scripts/seed.go --source mangadex --top 100
```

---

## 8. Rủi ro & cách giảm

| Rủi ro | Giảm thiểu |
|---|---|
| MangaDex/MAL rate limit, ban IP | `rate.Limiter` mỗi client + cache response 10 phút trong Mongo |
| Aggregator merge sai vì title khác nhau | ưu tiên match qua `external_ids`, fallback fuzzy match (Levenshtein) ≥ 0.85 |
| WebSocket leak goroutine | đảm bảo `defer close(send)`, ping timeout, unit test bằng `httptest` + `websocket.Dial` |
| gRPC service down → cả gateway hỏng | health check + circuit breaker đơn giản (`sony/gobreaker`) |
| JWT secret leak | đọc từ env, không commit, rotate được |
| Poller spam external API | chỉ poll những manga có ≥1 subscriber, chia batch |
| Mongo không có transaction giữa nhiều collection | dùng replica set 1 node trong compose để bật transaction khi cần (vd: tạo user + prefs) |

---

## 9. Checklist demo cuối khóa

- [ ] `docker compose up` 1 lệnh chạy hết.
- [ ] Postman collection demo: register → login → list manga → CRUD → reading status → stats.
- [ ] `wscat` demo: 2 client subscribe room `manga:one-piece`, trigger import chapter mới → cả 2 nhận được.
- [ ] `grpcurl` demo: gọi trực tiếp `catalog-svc.ListManga` để chứng minh microservice tách rời.
- [ ] Slide kiến trúc + mapping 4 lab → 4 layer.
