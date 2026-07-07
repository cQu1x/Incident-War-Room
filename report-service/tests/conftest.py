import sys
from pathlib import Path

import pytest


PROJECT_ROOT = Path(__file__).resolve().parents[1]

if str(PROJECT_ROOT) not in sys.path:
    sys.path.insert(0, str(PROJECT_ROOT))


@pytest.fixture(autouse=True)
def isolate_external_services(request, monkeypatch):
    if request.node.get_closest_marker("integration"):
        return

    monkeypatch.setenv("S3_ENABLED", "false")
    monkeypatch.delenv("DEEPSEEK_API_KEY", raising=False)
