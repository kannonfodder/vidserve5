# AI Coding Agent Instructions for vidserve5

These instructions help AI agents quickly contribute to this Go-based API/UI service. Focus on the `api-go/` folder; other paths are ancillary.

## Big Picture

- **Service type:** Monolithic Go HTTP server with Gorilla Mux and HTMX-friendly templating via `templ`.
- **Primary modules:**
  - `main.go` bootstraps env, DB, background worker, and routes.
  - `db/` provides PostgreSQL `pgxpool` connection via `InitDB()` (requires `DATABASE_URL`).
  - `auth/` manages user/session cookies; user IDs are UUIDs.
  - `components/` and `components/layout/` are `templ` UI templates compiled to Go by the `templ` tool.
  - `routes/` contains HTTP handlers grouped by feature (`login`, `register`, `feed`, `search`, `creators`, `rgp`).
  - `feedsvc/` implements the personal feed service and its background worker.
  - `api/redgifs/` integrates with Redgifs for video search.
- **Data flow:**
  1. HTTP requests routed in `main.go` → handlers in `routes/*/serve.go`.
  2. Handlers call services (`feedsvc`, `api/*`) and render `templ` components under `components/`.
  3. DB access via `pgxpool` from `db.InitDB()`; env drives connection.
  4. Background worker (`feedsvc.StartWorker`) fetches & stores feed items periodically.

## Developer Workflows

- **Run dev server (with templ auto-gen):** From `api-go/` use the watch script.
  - Windows PowerShell: see [api-go/watch.ps1](api-go/watch.ps1)
  - Command:
    ```powershell
    cd api-go
    templ generate --watch --proxy="http://localhost:8080" --cmd="go run ."
    ```
  - This compiles `.templ` files and proxies static refresh to the Go server.
- **Run server without templ:**
  ```powershell
  cd api-go
  go run .
  ```
- **Build & deploy (Docker/Kubernetes):**
  - Multi-arch image build & push then rollout:
    ```powershell
    cd api-go
    ./build.ps1
    ```
  - See [api-go/build.ps1](api-go/build.ps1). Requires Docker Buildx and `kubectl` access to the `vids` namespace.
- **DB setup:**
  - Env var `DATABASE_URL` must be present. Example in [api-go/FEED_SERVICE.md](api-go/FEED_SERVICE.md).
  - Apply migrations:
    ```bash
    psql whutbot < api-go/db/migrations.sql
    ```

## Conventions & Patterns

- **Routing:** Defined in `main.go` using Gorilla Mux.
  - Examples:
    - `/health` returns `200 OK`.
    - `/login` supports `GET` (render) and `POST` (authenticate) in `routes/login/serve.go`.
    - `/register` supports `GET` and `POST` in `routes/register/serve.go`.
    - `/creators/{username}` served by `routes/creators/serve.go`.
    - `/search` served by `routes/search/serve.go`.
    - `/rgp/*` prefix handled by `routes/rgp/serve.go`.
    - `/files/*` serves from local `files/` directory.
- **Templates (`templ`):**
  - `.templ` source files live under `components/` and `components/layout/`.
  - Generated `.go` files (e.g., `login_templ.go`) co-locate next to templates.
  - Use `Render(ctx, w)` to write responses (see `layout.Root("Kannonfoundry", layout.Search(user))`).
- **Auth:** `auth.IsLoggedIn(r)` returns user info for rendering; user ID persisted in cookies. When creating users, use UUIDs.
- **Feed service:**
  - Subscription types: `tag` or `creator`.
  - Cursor deduplication via `feed_subscriptions.last_video_id`.
  - Worker runs every ~10 minutes; starts on boot via `feedsvc.StartWorker(dbPool)`.
  - Key APIs: `CreateSubscription`, `DeleteSubscription`, `ListUserSubscriptions`, `GetUserFeed`.
- **DB access:**
  - Always obtain `*pgxpool.Pool` from `db.InitDB()` and pass to services/routes.
  - Do not create ad-hoc connections; reuse `dbPool`.

## External Integrations

- **Redgifs API:** `api/redgifs/redgifs.go` handles search and creator queries; feed fetcher relies on it.
- **HTMX-friendly UI:** Routes render `templ` components; keep responses SSR-first, enhance progressively.
- **Docker/K8s:** Image `docker.kannonfoundry.dev/api-go` and deployment `api-go` in namespace `vids`.

## Practical Examples

- **Render home with search:** In `main.go`, root handler:
  ```go
  user := auth.IsLoggedIn(r)
  layout.Root("Kannonfoundry", layout.Search(user)).Render(r.Context(), w)
  ```
- **Add a new route:**
  - Create `routes/feature/serve.go` with `Serve(w, r)`.
  - Register in `main.go`: `r.PathPrefix("/feature/").Handler(http.StripPrefix("/feature/", http.HandlerFunc(feature.Serve)))`.
- **Use feed service in a handler:**
  ```go
  items, _ := feedsvc.GetUserFeed(dbPool, userID, 20, 0)
  // render items via a templ component
  ```

## Gotchas

- **Env loading:** `.env` and `.key.env` loaded at startup; missing `DATABASE_URL` causes fatal init failure.
- **Windows dev:** Prefer running `watch.ps1`; it wires `templ` + `go run .`. If it fails, run `go run .` separately and `templ generate --watch` in another terminal.
- **Files storage:** `/new` creates a GUID-named file in `files/` and redirects; ensure `files/` exists.

## Where to Look

- Startup & wiring: [api-go/main.go](api-go/main.go)
- DB: [api-go/db/db.go](api-go/db/db.go), [api-go/db/migrations.sql](api-go/db/migrations.sql)
- Feed service: [api-go/FEED_SERVICE.md](api-go/FEED_SERVICE.md), [api-go/feedsvc](api-go/feedsvc)
- Templ components: [api-go/components](api-go/components)
- Routes: [api-go/routes](api-go/routes)
- Redgifs API: [api-go/api/redgifs](api-go/api/redgifs)
- Dev scripts: [api-go/watch.ps1](api-go/watch.ps1), [api-go/build.ps1](api-go/build.ps1)

If any of these sections are missing details (e.g., specific handler behavior, auth cookie format, or Redgifs client config), tell me what you need clarified and I’ll refine this guide.
