# Incident War Room

Incident War Room is a self-hosted incident management platform that integrates with Telegram, monitoring systems, a web dashboard, AI summarization, PDF reports, and S3-compatible object storage.

The platform helps engineering teams run production incident response directly from Telegram. It creates a dedicated Telegram Topic for every incident, records the discussion as a structured timeline, tracks severity changes, stores media evidence, generates AI-assisted summaries and PDF reports, and keeps incident history available in a dashboard.

The platform consists of three main parts:

* **Incident Service (Go)** - handles Telegram integration, incident lifecycle management, monitoring integrations, timeline tracking, severity history, media validation, dashboard authentication, data persistence, and communication with external services.
* **Report Service (Python)** - generates AI-enhanced PDF reports, adds incident statistics such as close time and duration, integrates with object storage, and supports both S3 and inline report delivery modes.
* **Dashboard (React)** - provides a web view of incidents, timelines, media, reports, and incident history scoped to the Telegram chat that requested access.

## Key Features

* Telegram Topics for isolated incident discussions
* Manual incident creation from Telegram
* Automatic incident creation from Prometheus Alertmanager
* Duplicate active incident detection with automatic numeric suffixes
* Incident reopen flow with preserved timeline history
* Chronological incident timeline
* Severity change history in Telegram, Telegraph, Dashboard, and PDF reports
* Support for multiple media attachments in timeline events
* Media validation before upload or timeline persistence
* AI-generated incident title and summary
* PDF report generation with closed time and incident duration
* S3-compatible report and media storage with fallback behavior
* Dashboard authentication via personal JWT links from Telegram
* Chat-scoped Dashboard access so users only see incidents from their Telegram chat
* PostgreSQL-based persistence
* Docker-based deployment

## Example Workflow

1. An incident is created manually from Telegram or automatically by Alertmanager.
2. A dedicated Telegram Topic is created for the incident.
3. Engineers discuss the incident inside the Topic while the system records timeline events.
4. The team can attach screenshots, logs, files, or other media to timeline messages.
5. Severity changes are recorded as timeline events and shown in Telegram, Telegraph, Dashboard, and the final PDF.
6. AI generates or improves the incident title and summary.
7. If another active incident is created with the same name, the platform keeps both incidents unique by adding a numeric suffix such as `-2` or `-3`.
8. When the incident is resolved, the Topic is closed or removed and the incident receives a `closed_at` timestamp.
9. A PDF report is generated with the timeline, summary, statistics, close time, duration, severity history, and media references.
10. If S3 is configured, report and media files are stored in object storage. If S3 report upload is unavailable, the report can still be delivered directly.
11. The incident remains available through the Dashboard for future review.
12. A closed incident can be reopened; the existing timeline is preserved, a new Telegram Topic is created, and an `INCIDENT_REOPENED` event is added.

The platform is designed for engineering teams that already use Telegram as their primary communication tool and want a lightweight incident management solution that can be tested quickly or deployed in their own infrastructure.

## Quick Start

Use one of the two modes:

* **Hosted Bot** - use the team deployment without running any infrastructure yourself.
* **Self-Hosted Mode** - run your own bot, database, report service, dashboard, and optional integrations.

### Hosted Bot

Hosted Bot is the fastest way to try Incident War Room. Use this mode if you do not want to self-host, do not have infrastructure resources, or want to test the bot before deploying it yourself.

Use the public Telegram bot link shared by the Incident War Room team. The hosted deployment already includes the backend services, report generation, Dashboard, and storage configuration, so you only need to add the bot to your Telegram group and start working with incidents.

#### Add the bot to a Telegram group

1. Open the public Incident War Room bot link in Telegram.
2. Click **Start** to initialize the bot in a direct chat.
3. Add the bot to the target Telegram group.
4. Convert the group to a supergroup if Telegram has not done this automatically.
5. Enable **Topics** in the group settings.
6. Promote the bot to administrator.
7. Give the bot permissions to manage topics and send messages. If media collection is enabled, also allow it to read and process group messages.
8. Open the bot menu or send the first incident creation action in the group to verify that the bot is active.

#### Basic incident flow

1. Create an incident from the bot command or menu in the group.
2. The bot creates a dedicated Telegram Topic for that incident.
3. Work inside the incident Topic. Messages, decisions, status updates, severity changes, and media attachments become part of the timeline.
4. Use severity controls when the impact changes. Each severity update is recorded as a `SEVERITY_CHANGED` timeline event.
5. Attach screenshots, logs, or other media directly in the Topic. The bot validates media and stores it when storage is enabled.
6. Open the Dashboard with `/dashboard` when you need a web view.
7. Resolve the incident when the response is finished.
8. The bot generates a PDF report and sends it back to Telegram.
9. Reopen a closed incident if the problem returns. The bot preserves the old timeline, creates a new Topic, and adds a reopen event.

#### Dashboard access

Use:

```text
/dashboard
```

The bot generates a personal Dashboard link. The link contains a JWT bound to the Telegram chat ID, so the Dashboard only shows incidents from the chat where the link was requested. If the session expires, request a new link with `/dashboard`.

#### What the hosted bot can do

* Create incidents in Telegram.
* Create one Telegram Topic per incident.
* Record a chronological timeline.
* Track severity changes.
* Accept multiple media attachments per timeline event when media storage is configured.
* Generate AI-assisted titles and summaries.
* Generate PDF reports with close time, duration, timeline, severity history, and media references.
* Show incidents in the public Dashboard.
* Prevent duplicate active incident names by adding numeric suffixes.
* Reopen closed incidents while preserving history.
* Create incidents from Alertmanager if the hosted deployment has been connected to a monitoring source.

### Self-Hosted Mode

Self-Hosted Mode lets you run the full Incident War Room stack on your own machine or server. You provide your own Telegram bot token, database, optional S3-compatible storage, optional Alertmanager integration, and optional public Dashboard URL.

Use this mode when you need full control over data, infrastructure, tokens, storage, or integrations.

#### Prerequisites

* Docker with the Compose plugin
* A Telegram bot token from [@BotFather](https://t.me/BotFather)
* A Telegram group with Topics enabled
* A DeepSeek API key for AI title and summary generation
* Optional: S3-compatible object storage
* Optional: Prometheus Alertmanager
* Optional: a public domain or reverse proxy for Dashboard links

#### 1. Configure Telegram

1. Create a bot in [@BotFather](https://t.me/BotFather).
2. Copy the bot token into `BOT_TOKEN`.

#### 2. Configure environment variables

Create a root `.env` file from the template:

```bash
cp .env.example .env
```

Fill in the required values:

| Variable | Description |
| --- | --- |
| `BOT_TOKEN` | Telegram bot token from @BotFather |
| `POSTGRES_DB` | PostgreSQL database name |
| `POSTGRES_USER` | PostgreSQL username |
| `POSTGRES_PASSWORD` | PostgreSQL password |
| `DEEPSEEK_API_KEY` | API key for AI title and summary generation |

Compose sets these service addresses automatically. Keep the template values unless you know you need a custom network layout:

| Variable | Description |
| --- | --- |
| `POSTGRES_HOST` | PostgreSQL hostname inside Docker Compose |
| `REPORT_SERVICE_URL` | Report Service URL inside Docker Compose |

Optional variables:

| Variable | Default | Description |
| --- | --- | --- |
| `JWT_SECRET` | empty | HS256 signing secret for Dashboard links. If empty, Dashboard links are not issued. |
| `DASHBOARD_URL` | `https://incident-war-room.ru` | Base URL used when the bot generates Dashboard links. Set this to your deployed frontend URL. |
| `DASHBOARD_TOKEN_TTL` | `168h` | Dashboard JWT lifetime. After expiration, users must request a new link with `/dashboard`. |
| `TELEGRAPH_ACCESS_TOKEN` | empty | Token for live timeline Telegraph pages. If empty, an anonymous account is created on first use. |
| `S3_ENABLED` | `false` | Enables or disables S3-compatible storage for media and reports. See the S3 section below. |
| `S3_ENDPOINT_URL` | empty | S3-compatible endpoint URL. For AWS S3, use the AWS endpoint. For MinIO, use your MinIO URL. |
| `S3_REGION` | empty | S3 region. Some S3-compatible providers require any non-empty region. |
| `S3_BUCKET_NAME` | empty | Bucket used for uploaded media and reports. |
| `S3_ACCESS_KEY` | empty | S3 access key. |
| `S3_SECRET_KEY` | empty | S3 secret key. |
| `S3_PUBLIC_BASE_URL` | empty | Public base URL used to build links to uploaded objects. |
| `ALERT_CHAT_ID` | empty | Telegram chat ID where Alertmanager-created incidents should be opened. |
| `ALERTMANAGER_WEBHOOK_TOKEN` | empty | Shared token used to authenticate Alertmanager webhooks. |
| `DEEPSEEK_MODEL` | `deepseek-v4-flash` | AI model override. |
| `HTTP_ADDR` | `:8080` | Incident Service listen address. |
| `CORS_ALLOWED_ORIGIN` | `*` | Allowed CORS origin for the frontend. In production, set this to your Dashboard origin. |

#### 3. Understand S3 modes and fallback behavior

`S3_ENABLED=false`

* Object storage is disabled.
* PDF reports are delivered directly by the Report Service when possible.
* Timeline media attachments are not persisted to S3.
* Media that requires object storage can be rejected because there is no durable public storage location.
* This mode is useful for local tests and minimal deployments.

`S3_ENABLED=true`

* The platform uploads reports and supported media to S3-compatible storage.
* `S3_ENDPOINT_URL`, `S3_REGION`, `S3_BUCKET_NAME`, `S3_ACCESS_KEY`, and `S3_SECRET_KEY` must be valid.
* `S3_PUBLIC_BASE_URL` should point to the public URL prefix for the bucket or reverse proxy so Telegram, Telegraph, Dashboard, and PDF reports can reference uploaded files.
* If report upload to S3 fails or S3 is misconfigured, the Report Service can fall back to inline PDF delivery instead of losing the report.
* Media fallback is stricter: if media storage is required and S3 is unavailable, media attachments may be rejected or omitted because they cannot be served later from the timeline, Dashboard, or PDF.

Recommended production setup:

```env
S3_ENABLED=true
S3_ENDPOINT_URL=https://s3.example.com
S3_REGION=us-east-1
S3_BUCKET_NAME=incident-war-room
S3_ACCESS_KEY=...
S3_SECRET_KEY=...
S3_PUBLIC_BASE_URL=https://cdn.example.com/incident-war-room
```

#### 4. Start the backend stack

```bash
docker compose up --build -d
```

or:

```bash
make up
```

This starts PostgreSQL, applies migrations, starts the Report Service, and starts the Incident Service. Migrations in `incident-service/migration/*_up.sql` run automatically before the Incident Service starts.

Host ports:

| Service | Port |
| --- | --- |
| Incident Service API | `8080` |
| PostgreSQL | `5432` |
| Report Service | `8000` inside the Docker network only |

#### 5. Check the services

```bash
docker compose ps
docker compose logs -f
```

or:

```bash
make ps
make logs
```

#### 6. Run the Dashboard

The React frontend is not part of `docker-compose.yml`; run it separately:

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

Frontend variables:

| Variable | Default | Description |
| --- | --- | --- |
| `VITE_INCIDENT_API_BASE` | `http://localhost:8080` | Incident Service API base URL |
| `VITE_REPORT_API_BASE` | `http://localhost:8000` | Report Service base URL |

For production, build the frontend and serve it behind your web server or reverse proxy:

```bash
npm run build
```

Set `DASHBOARD_URL` in the backend `.env` to the public URL where this frontend is available.

#### 7. Connect Alertmanager (optional)

Set:

```env
ALERT_CHAT_ID=<telegram-chat-id>
ALERTMANAGER_WEBHOOK_TOKEN=<shared-secret>
```

Then configure Prometheus Alertmanager to send webhook requests to the Incident Service. Use the same shared token expected by the deployment. Alerts create incidents automatically in `ALERT_CHAT_ID`.

#### 8. Use the self-hosted bot

1. Add your bot to a Telegram group.p.
2. Enable Topics in the group.
3. Promote the bot to administrator.
4. Give the bot permission to manage topics and send messages.
5. If the bot must collect media and timeline messages, make sure Telegram privacy settings and group permissions allow the bot to see the required messages.
6. Open the bot menu or run the first incident creation action to verify that the bot is running.
7. Create an incident from the bot command or menu.
8. Work inside the generated Topic.
9. Use `/dashboard` to generate a personal Dashboard link if `JWT_SECRET` and `DASHBOARD_URL` are configured.
10. Resolve the incident and check that the PDF report is delivered.
11. If S3 is enabled, verify that media and reports are available through the configured public base URL.

#### 9. Stop or reset

Stop without deleting data:

```bash
docker compose down
```

or:

```bash
make down
```

Stop and delete the PostgreSQL volume:

```bash
docker compose down -v
```

or:

```bash
make clean
```

`down -v` deletes the `postgres_data` volume and all incident data. Use plain `down` to stop without data loss.

## Operations Notes

* Keep `BOT_TOKEN`, `DEEPSEEK_API_KEY`, `JWT_SECRET`, and S3 credentials secret.
* Use a strong `JWT_SECRET` in every production deployment.
* Restrict `CORS_ALLOWED_ORIGIN` to the real Dashboard origin in production.
* Make the Dashboard URL public if users need to open links generated by `/dashboard`.
* Use S3-compatible storage in production if media attachments and durable report links are required.
* Verify that the bot has Telegram permissions to manage Topics before testing incident creation.
