from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


def test_health_endpoint() -> None:
    response = client.get("/api/v1/health")
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


def test_capabilities_endpoint_contains_core_modules() -> None:
    response = client.get("/api/v1/system/capabilities")
    assert response.status_code == 200

    data = response.json()
    names = {capability["name"] for capability in data["capabilities"]}

    assert data["platform"] == "Kyver"
    assert "multi-agent" in names
    assert "rag-knowledge-retrieval" in names
    assert "plugin-system" in names
