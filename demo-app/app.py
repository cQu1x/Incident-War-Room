"""Demo target for the incident war room demo.

Exposes a single Prometheus gauge `demo_app_error_rate` that you can flip
between healthy (0) and broken (1) from a tiny web page. Prometheus scrapes
this app; when the gauge crosses the alert threshold an alert fires and an
incident is opened in the war room.
"""

from fastapi import FastAPI, Response
from fastapi.responses import HTMLResponse
from prometheus_client import Gauge, generate_latest, CONTENT_TYPE_LATEST

app = FastAPI(title="demo-app")

error_rate = Gauge("demo_app_error_rate", "Simulated error rate of the demo app (0 = healthy, 1 = broken)")

# Current value, mirrored into the gauge. Kept as plain state so the page can
# render it without touching prometheus_client internals.
state = {"value": 0.0}
error_rate.set(state["value"])


def _set(value: float) -> None:
    state["value"] = value
    error_rate.set(value)

PAGE = """<!doctype html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>demo-app</title>
  <style>
    body {{ font-family: system-ui, sans-serif; max-width: 480px; margin: 60px auto; text-align: center; }}
    h1 {{ font-size: 1.4rem; }}
    .state {{ font-size: 2rem; margin: 24px 0; }}
    .ok {{ color: #16a34a; }}
    .bad {{ color: #dc2626; }}
    button {{ font-size: 1.1rem; padding: 14px 24px; margin: 8px; border: 0; border-radius: 10px; cursor: pointer; color: #fff; }}
    .break {{ background: #dc2626; }}
    .fix {{ background: #16a34a; }}
  </style>
</head>
<body>
  <h1>demo-app — управляемый таргет</h1>
  <div class="state {cls}">{label}</div>
  <p>error_rate = <b>{value}</b></p>
  <button class="break" onclick="act('/break')">🔥 Сломать</button>
  <button class="fix" onclick="act('/fix')">✅ Починить</button>
  <script>
    async function act(path) {{
      await fetch(path, {{ method: 'POST' }});
      location.reload();
    }}
  </script>
</body>
</html>"""


def _render() -> str:
    broken = state["value"] > 0.5
    return PAGE.format(
        cls="bad" if broken else "ok",
        label="🔥 СЛОМАНО" if broken else "✅ Работает",
        value=state["value"],
    )


@app.get("/", response_class=HTMLResponse)
def index() -> str:
    return _render()


@app.post("/break")
def do_break() -> dict:
    _set(1.0)
    return {"error_rate": state["value"]}


@app.post("/fix")
def do_fix() -> dict:
    _set(0.0)
    return {"error_rate": state["value"]}


@app.get("/metrics")
def metrics() -> Response:
    return Response(generate_latest(), media_type=CONTENT_TYPE_LATEST)
