# TCP & UDP Modules

Two standalone binaries added to satisfy the network-protocol requirements
of the project spec. Both run independently and integrate with the gateway
through a tiny internal HTTP publish API — there is **no** gRPC dependency
between gateway and these two services, which keeps them lightweight and
easy to demo.

```
┌──────────┐   HTTP    ┌────────────┐    TCP    ┌───────────────┐
│ Gateway  │──────────▶│  tcp-svc   │──────────▶│ tcp-client(s) │
│          │  /publish │  :9000 tcp │ broadcast │ multi-device  │
│          │           │  :9100 http│           └───────────────┘
│          │   HTTP    ┌────────────┐    UDP    ┌───────────────┐
│          │──────────▶│  udp-svc   │──────────▶│ udp-client(s) │
│          │  /publish │  :9001 udp │ broadcast │ notifier apps │
└──────────┘           │  :9101 http│           └───────────────┘
                       └────────────┘
```

## TCP Progress Sync (`cmd/tcp-server`)

| Port | Protocol | Purpose |
|---|---|---|
| 9000 | TCP   | Raw TCP, newline-delimited JSON. Clients connect, send `subscribe`, then receive every `progress_update` for that user. |
| 9100 | HTTP  | `POST /publish` so the gateway can fan out new progress events. |

### Client protocol (newline-delimited JSON over TCP)

Client → server (after dialling):
```json
{"type":"subscribe","user_id":"665a1d8f3c0e2a4b1c8d9e10"}
```

Server → client (any time):
```json
{"type":"system","content":"connected"}
{"type":"subscribed","content":"665a1d8f3c0e2a4b1c8d9e10"}
{"type":"progress_update","user_id":"…","manga_id":"…","chapter":47,"status":"reading","timestamp":1730000000}
```

Optional client heartbeat: `{"type":"ping"}` → server replies `{"type":"pong"}`.

### How events get there

`PUT /me/reading/:mangaId` on the gateway publishes a `ProgressEvent` to
`tcp-svc` via the internal HTTP `POST /publish`. The TCP server then writes
the event to every connection currently subscribed to the matching `user_id`.

## UDP Notification (`cmd/udp-server`)

| Port | Protocol | Purpose |
|---|---|---|
| 9001 | UDP   | Register / heartbeat / receive notification datagrams. |
| 9101 | HTTP  | `POST /publish` so the gateway / poller can trigger broadcasts. |

### Client protocol (single-datagram JSON over UDP)

Client → server:
```json
{"type":"REGISTER","client_id":"desktop-01"}
{"type":"HEARTBEAT","client_id":"desktop-01"}
{"type":"UNREGISTER","client_id":"desktop-01"}
{"type":"ACK","msg_id":"f3a1b2c4..."}   // optional delivery confirmation
```

Server → client:
```json
{"type":"new_chapter","msg_id":"f3a1b2c4...","manga_id":"…","chapter":1101,"message":"Chapter 1101 released!","timestamp":1730000000}
```

### How events get there

* `MangaPoller` detects a "new" chapter for any subscribed manga (see
  `internal/core/poller/manga_poller.go`) and POSTs to `udp-svc`.
* `POST /admin/notify` (existing admin endpoint) also forwards to `udp-svc`,
  so manual triggering from Postman is easy.

The UDP registry has a TTL (default 90 s) — clients that stop heartbeating
are GC'd automatically.

## Running

### Locally (multi-terminal, no docker)

```powershell
# terminal 1
go run ./cmd/tcp-server

# terminal 2
go run ./cmd/udp-server

# terminal 3 — gateway needs to know where to publish
$env:TCP_PUBLISH_URL = "http://localhost:9100/publish"
$env:UDP_PUBLISH_URL = "http://localhost:9101/publish"
go run ./cmd/gateway
```

### With docker compose

```powershell
docker compose -f deploy/docker-compose.yml up --build
```

The compose file already wires the URLs and exposes the relevant ports.

## Demo clients

```powershell
# Subscribe to a user's progress feed
go run ./cmd/tcpclient -addr=localhost:9000 -user=<user_id_hex>

# Register for chapter notifications (with ACK enabled)
go run ./cmd/udpclient -server=localhost:9001 -id=desktop-01 -ack
```

A scripted version that spawns 2 of each client and drives a publish through
the gateway is at `scripts/demo-tcp-udp.ps1`.

## Manual verification with Postman

Postman cannot talk raw TCP/UDP, but it can drive the gateway HTTP endpoints
that ultimately publish to tcp-svc / udp-svc:

* `PUT /me/reading/:mangaId` — triggers TCP broadcast.
* `POST /admin/notify` — triggers UDP broadcast.

Open `tcpclient` / `udpclient` (or netcat) in separate terminals to observe
the broadcasts arriving.

## Internal publish endpoints (debug)

If you want to bypass the gateway and poke the broadcast services directly:

```http
POST http://localhost:9100/publish
Content-Type: application/json

{ "user_id":"…", "manga_id":"…", "chapter":47, "status":"reading" }
```

```http
POST http://localhost:9101/publish
Content-Type: application/json

{ "manga_id":"…", "manga_title":"One Piece", "chapter":1101, "message":"Chapter 1101!" }
```

Set `INTERNAL_TOKEN` in both env and request header `X-Internal-Token` to
gate these endpoints in production.
