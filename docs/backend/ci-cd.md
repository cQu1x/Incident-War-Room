# CI/CD

This document describes the continuous integration and deployment pipeline for
Incident War Room. Both pipelines are implemented as GitHub Actions workflows:

| Workflow | File | Trigger | Purpose |
|----------|------|---------|---------|
| CI | [`.github/workflows/ci.yml`](../../.github/workflows/ci.yml) | every pull request and every push to `main` | build, lint, test and package all services |
| CD | [`.github/workflows/cd.yml`](../../.github/workflows/cd.yml) | after CI succeeds on `main`, or manually | deploy the packaged images to the production VM |

The guiding principle is **build once, deploy the same artifact**: images are
built and tested in CI, published to the GitHub Container Registry (GHCR), and
CD deploys those exact images instead of rebuilding on the server.

```
 pull request ─▶ CI (build + test, no publish)

 push to main ─▶ CI (build + test + push images to GHCR)
                     │  on success (workflow_run)
                     ▼
                  CD ─▶ pull images on the VM ─▶ up -d ─▶ health check
```

---

## Continuous Integration (`ci.yml`)

CI runs on every pull request and on every push to `main`. It has three
independent jobs that run in parallel.

Two workflow-level settings apply to all jobs:

- `permissions: contents: read` — the default token is least-privilege; the
  `docker` job widens this to `packages: write` only for itself.
- `concurrency` — a run is grouped by git ref, and superseded **pull-request**
  runs are cancelled (`cancel-in-progress` is true only for PRs). Runs on
  `main` are never cancelled, so every merged commit is fully validated.

### Job: `incident-service` (Go)

Runs in `incident-service/` with Go resolved from `go.mod` and the module cache
keyed on `go.sum`.

| Step | Command | Fails when |
|------|---------|------------|
| gofmt | `test -z "$(gofmt -l ./internal/ ./cmd/)"` | any file is not gofmt-formatted |
| go vet | `go vet ./...` | the vet analyzer reports a problem |
| go build | `go build ./...` | the module does not compile |
| go test | `go test ./...` | any test fails |

### Job: `report-service` (Python)

Runs in `report-service/` on Python 3.12 with the pip cache keyed on
`requirements.txt`.

| Step | Command | Notes |
|------|---------|-------|
| Install dependencies | `pip install -r requirements.txt` | |
| pytest | `pytest -m "not integration"` | integration tests (real S3 / network) are excluded in CI |

Tests are **hermetic**: an autouse fixture in
[`report-service/tests/conftest.py`](../../report-service/tests/conftest.py)
forces `S3_ENABLED=false` and removes `DEEPSEEK_API_KEY` for non-integration
tests, so they never contact S3 or the DeepSeek API and do not depend on a
local `.env`. Integration-marked tests are left untouched and are only run
deliberately (they are deselected here by `-m "not integration"`).

### Job: `docker build`

This job validates and packages the Docker images. It has
`permissions: packages: write` so it can publish to GHCR.

Image coordinates are derived from two environment variables:

- `IMAGE_PREFIX` = `ghcr.io/<owner>` (lower-cased in the *Normalize image
  prefix* step, because GHCR requires lowercase names).
- `IMAGE_TAG` = the commit SHA (`github.sha`).

Steps:

1. **Provide env for compose** — `cp .env.example .env` so Compose can
   interpolate variables.
2. **Validate compose file** — `docker compose config -q` catches syntax and
   interpolation errors.
3. **Build images** — `docker compose build` builds all four buildable
   services. Because each service declares both `build:` and `image:`, the
   built images are tagged as `ghcr.io/<owner>/<service>:<sha>`.
4. **Log in to GHCR** — only on push to `main`, using the built-in
   `GITHUB_TOKEN`.
5. **Push images** — only on push to `main`. Pushes `incident-service`,
   `report-service`, `frontend` and `demo-app`, each tagged with both the
   commit SHA and `latest`.

On **pull requests** steps 4–5 are skipped, so a PR only proves that everything
builds — nothing is published.

---

## Container images

The four application images are built from the `build:` sections in
[`docker-compose.yml`](../../docker-compose.yml) and published to GHCR:

| Service | Source | Published image |
|---------|--------|-----------------|
| incident-service | `./incident-service` | `ghcr.io/<owner>/incident-service` |
| report-service | `./report-service` | `ghcr.io/<owner>/report-service` |
| frontend | `./frontend` | `ghcr.io/<owner>/frontend` |
| demo-app | `./demo-app` | `ghcr.io/<owner>/demo-app` |

Each of these services declares both keys in Compose:

```yaml
image: ${IMAGE_PREFIX:-ghcr.io/incidentwarroom}/incident-service:${IMAGE_TAG:-latest}
build: ./incident-service
```

This single declaration serves every context:

- **Local development** — `docker compose up --build` builds and tags the image
  locally; the defaults (`ghcr.io/incidentwarroom/...:latest`) mean nothing has
  to be set and no registry is contacted.
- **CI** — `docker compose build` tags the images, `docker compose push`
  publishes them under `IMAGE_PREFIX` / `IMAGE_TAG`.
- **CD** — `docker compose pull` fetches the exact tag, `docker compose up -d`
  runs it without rebuilding.

The remaining services (`postgres`, `caddy`, `frpc`, `prometheus`,
`alertmanager`, and the one-shot `migrate`) use upstream public images and are
not built or published by us.

### Image tags

| Tag | Meaning |
|-----|---------|
| `<commit-sha>` | immutable, points at exactly the code that produced it |
| `latest` | moves to the newest build on `main` |

CD deploys the **commit SHA** when it is triggered by CI (so the deployed image
matches the tested commit exactly) and falls back to `latest` for manual runs.

---

## Continuous Deployment (`cd.yml`)

### Trigger and gating

CD is triggered by `workflow_run`: it starts **after the CI workflow completes
on `main`**, and the `deploy` job only runs when CI concluded successfully:

```yaml
if: ${{ github.event_name == 'workflow_dispatch' || github.event.workflow_run.conclusion == 'success' }}
```

This ordering guarantees the images already exist in GHCR before CD tries to
pull them, and that a red CI never deploys. CD can also be started manually via
**workflow_dispatch**.

A `concurrency` group (`deploy-production`, `cancel-in-progress: false`)
serialises deploys so two releases never run at once.

### Runner

CD runs on a **self-hosted runner** labelled `[self-hosted, incident-vm]`,
i.e. the production VM itself. The VM sits behind NAT and cannot accept inbound
connections, so deployment is pull-based: the runner polls GitHub for jobs.
See [`docs`](../) and the deployment notes for the runner setup.

### Steps

1. **Log in to GHCR** — `docker login ghcr.io` with the built-in
   `GITHUB_TOKEN` (`packages: read`) so private images can be pulled.
2. **Deploy** (in `${{ secrets.DEPLOY_PATH }}`):
   ```sh
   git fetch --prune origin
   git reset --hard "$TARGET_REF"   # the tested commit, or origin/main for manual runs
   docker compose pull
   docker compose up -d
   ```
   The repository checkout is still required because Compose needs the
   `docker-compose.yml`, `Caddyfile`, `frpc.toml`, monitoring configs and
   migration scripts from the working tree — only the application **images**
   come from GHCR.
3. **Health check** — waits up to 120 s for `report-service` to report
   `healthy` (it defines a `/health` healthcheck in Compose) and asserts that
   `incident-service` is `running`. On failure it prints `docker compose ps`
   and the last 80 log lines and exits non-zero, so a broken release fails the
   deploy instead of silently replacing a working stack.
4. **Prune old images** — `docker image prune -f`, run only after the stack is
   confirmed healthy.

### Environment variables

| Variable | Value | Purpose |
|----------|-------|---------|
| `IMAGE_PREFIX` | `ghcr.io/<owner>` (lower-cased) | registry namespace to pull from |
| `IMAGE_TAG` | `workflow_run.head_sha`, else `latest` | which image tag to deploy |
| `TARGET_REF` | `workflow_run.head_sha`, else `origin/main` | which commit to check out for config/compose |

### Required secrets

| Secret | Used for |
|--------|----------|
| `DEPLOY_PATH` | working directory of the checkout on the VM |
| `GITHUB_TOKEN` | provided automatically by Actions; used to pull from GHCR |

Application secrets (`DEEPSEEK_API_KEY`, S3 credentials, `FRP_AUTH_TOKEN`,
Postgres credentials, the Telegram token, …) live in the `.env` file on the VM
and are **not** managed by these workflows. See
[`.env.example`](../../.env.example) for the full list.

---

## Release flow, end to end

1. Open a pull request → CI builds and tests everything; nothing is published.
2. Merge to `main` → CI runs again and, on success, pushes images tagged with
   the merge commit SHA (and `latest`) to GHCR.
3. CD starts automatically after that CI run succeeds, pulls the SHA-tagged
   images on the VM, brings the stack up and health-checks it.
4. If the health check fails, the deploy is marked failed and the previous
   containers that could not be replaced keep running; investigate with
   `docker compose ps` / `docker compose logs` on the VM.

## Rollback

To roll back, redeploy a previous image tag. Either:

- re-run CD via **workflow_dispatch** after resetting `main` to the previous
  commit, or
- on the VM, set `IMAGE_TAG` to a known-good commit SHA and run
  `docker compose pull && docker compose up -d` manually.

Because every build is tagged with its commit SHA, any previously built commit
can be redeployed as long as its images are still present in GHCR.

## Running the checks locally

Before opening a PR you can reproduce each CI job:

```sh
# incident-service (Go)
cd incident-service
test -z "$(gofmt -l ./internal/ ./cmd/)"
go vet ./... && go build ./... && go test ./...

# report-service (Python)
cd report-service
pip install -r requirements.txt
pytest -m "not integration"

# docker
cp .env.example .env
docker compose config -q
docker compose build
```
