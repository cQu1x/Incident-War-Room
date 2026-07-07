# Incident War Room

Incident War Room is a self-hosted incident management platform that integrates with Telegram and monitoring systems to simplify production incident response.

The platform automatically creates dedicated Telegram Topics for incidents, records investigation updates as a structured timeline, generates AI-powered incident reports, and stores reports in S3-compatible object storage. It can also integrate with Prometheus Alertmanager to automatically create incidents when monitoring detects a failure.

The platform consists of two core services:

* **Incident Service (Go)** — handles Telegram integration, incident lifecycle management, monitoring integrations, timeline tracking, data persistence, and communication with external services.
* **Report Service (Python)** — generates AI-enhanced PDF reports, integrates with object storage, and supports both S3 and inline report delivery modes.

## Key Features

* Telegram Topics for isolated incident discussions
* Automatic incident creation from Prometheus Alertmanager
* Manual incident creation directly from Telegram
* Chronological incident timeline
* AI-generated incident title and summary
* PDF report generation
* S3-compatible report storage with automatic fallback
* Incident history and web dashboard
* PostgreSQL-based persistence
* Docker-based deployment

## Example Workflow

1. An incident is created manually from Telegram or automatically by Alertmanager.
2. A dedicated Telegram Topic is created for the incident.
3. Engineers discuss the incident inside the Topic while the system records the timeline.
4. AI generates an incident title and summary.
5. The incident is resolved and the Topic is automatically removed.
6. A PDF report is generated and delivered to Telegram.
7. If S3 is configured, the report is stored in object storage; otherwise, it is delivered directly by the Report Service.
8. The incident remains available through the web dashboard for future investigation and analysis.

The platform is designed for engineering teams that already use Telegram as their primary communication tool and want a lightweight, self-hosted incident management solution that can be deployed in their own infrastructure and integrated with existing monitoring systems.

## Running with Docker

Requires Docker with the Compose plugin.

### 1. Set environment variables

All services read a single root `.env` file. Create it from the template:

```bash
cp .env.example .env
```

Fill in the values. Required:

| Variable | Description |
| --- | --- |
| `BOT_TOKEN` | Telegram bot token from @BotFather |
| `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` | Database name and credentials |
| `DEEPSEEK_API_KEY` | API key for AI title/summary generation |

`POSTGRES_HOST` and `REPORT_SERVICE_URL` are set automatically by Compose — leave the `.env` values as-is.

Optional (leave empty/default to disable the feature):

| Variable | Default | Enables |
| --- | --- | --- |
| `JWT_SECRET` | _(empty)_ | Web dashboard links (HS256 signing secret; no secret = no link issued) |
| `DASHBOARD_URL` | `https://incident-war-room.ru` | Base URL used in dashboard links |
| `DASHBOARD_TOKEN_TTL` | `168h` | Dashboard token lifetime |
| `TELEGRAPH_ACCESS_TOKEN` | _(empty)_ | Live timeline pages (empty = anonymous account created on first use) |
| `S3_ENABLED` + `S3_ENDPOINT_URL`, `S3_REGION`, `S3_BUCKET_NAME`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_PUBLIC_BASE_URL` | `false` | Media attachments in the timeline (off = media rejected) |
| `ALERT_CHAT_ID` + `ALERTMANAGER_WEBHOOK_TOKEN` | _(empty)_ | Incidents auto-created from Prometheus Alertmanager |
| `DEEPSEEK_MODEL` | `deepseek-v4-flash` | Overrides the AI model |
| `HTTP_ADDR` | `:8080` | Incident-service listen address |
| `CORS_ALLOWED_ORIGIN` | `*` | Allowed CORS origin for the frontend |

### 2. Start

```bash
docker compose up --build -d      # or: make up
```

This starts Postgres, applies migrations, then starts the report-service and incident-service. Migrations in `incident-service/migration/*_up.sql` run automatically before the incident-service — nothing to run by hand.

Host ports: incident-service API on `8080`, Postgres on `5432`. The report-service listens on `8000` inside the Docker network only.

### 3. Manage

```bash
docker compose ps                 # status         (make ps)
docker compose logs -f            # tail logs       (make logs)
docker compose down               # stop, keep DB   (make down)
docker compose down -v            # stop, WIPE DB volume (make clean)
```

`down -v` deletes the `postgres_data` volume and all incident data. Use plain `down` to stop without data loss.

## Frontend (web dashboard)

The dashboard is private: it renders only for someone holding a valid access
token. Operators get their personal link from the Telegram bot's `/dashboard`
command (and when an incident closes). The link carries a chat-scoped JWT; the
frontend stores it and sends it as a bearer token on every API call, and scopes
the incident list to that chat. Opening the site without a token shows a
"open it from the bot" screen.

### Local development

```bash
cd frontend
cp .env.example .env          # set VITE_INCIDENT_API_BASE to the incident-service URL
npm install
npm run dev                   # dev server (npm run build for a production bundle)
```

| Variable | Default | Description |
| --- | --- | --- |
| `VITE_INCIDENT_API_BASE` | `http://localhost:8080` | Incident-service API base URL |
| `VITE_REPORT_API_BASE` | `http://localhost:8000` | Report-service base URL |

To reach a dev incident-service you'll need a token: run `/dashboard` against
your bot and paste the `?token=…` from the link onto your dev URL once.

## Public deployment (domain + HTTPS)

The root `docker-compose.yml` includes a `frontend` (static React build served
by nginx) and a `caddy` reverse proxy. Caddy serves the frontend and proxies
`/api/*` to the incident-service under one domain, and obtains/renews a
Let's Encrypt TLS certificate automatically.

1. Point your domain's `A` record at the server and open ports `80` and `443`.
2. In the root `.env`, set (in addition to the required vars above):

   | Variable | Description |
   | --- | --- |
   | `SITE_DOMAIN` | Public domain Caddy serves and gets a cert for, e.g. `incident-war-room.ru` |
   | `PUBLIC_BASE_URL` | Public https URL of that domain — baked into the frontend as its API base |
   | `DASHBOARD_URL` | Same public https URL — used to build the `/dashboard` link |
   | `JWT_SECRET` | Non-empty random secret; without it the bot issues no dashboard link |

3. Start everything:

   ```bash
   docker compose up --build -d
   ```

4. In Telegram, run `/dashboard` and open the link — it lands on the domain with
   your token and loads the incident list over HTTPS.

`PUBLIC_BASE_URL` is inlined into the JS bundle at build time, so after changing
it rebuild the frontend (`docker compose up --build -d`).

## Automatic incidents from monitoring

`docker-compose.yml` ships a demo monitoring stack that opens incidents
automatically when monitoring detects a failure:

* **demo-app** (`:9000`) — a controllable target with a web page exposing a
  Prometheus gauge and two buttons: **🔥 Сломать** / **✅ Починить**.
* **prometheus** (`:9090`) — scrapes demo-app and evaluates the `HighErrorRate`
  rule (`demo_app_error_rate > 0.5` for 15s).
* **alertmanager** (`:9093`) — on a firing alert, POSTs to the incident-service
  webhook `POST /webhooks/alertmanager`, which opens an incident.

Required `.env` for this to work:

| Variable | Description |
| --- | --- |
| `ALERT_CHAT_ID` | Numeric id of the Telegram forum supergroup (bot must be admin with Manage Topics) where the auto-incident is opened. Without it the webhook returns "alert chat is not configured" |
| `ALERTMANAGER_WEBHOOK_TOKEN` | Bearer token that authenticates the webhook. **Must match** `credentials:` in `monitoring/alertmanager/alertmanager.yml` |

**Try it:** open the demo-app page (`http://<host>:9000`), press **🔥 Сломать**.
Within ~20–30s Prometheus fires `HighErrorRate`, Alertmanager calls the webhook,
and a new incident appears in the alert chat. Press **✅ Починить** to reset the
gauge.

> The webhook endpoint is only reachable inside the Docker network (Caddy exposes
> only `/api/*` and the frontend), so nothing external can post fake alerts. To
> click the demo button from your browser, open port `9000` on the host firewall.

## CI/CD

Every pull request and every push to `main` runs the CI workflow (Go and Python
build/lint/test plus a Docker build). On `main`, CI also publishes the service
images to GHCR, and the CD workflow then deploys those exact images to the
production VM and health-checks the stack.

See [`docs/backend/ci-cd.md`](docs/backend/ci-cd.md) for the full description.
