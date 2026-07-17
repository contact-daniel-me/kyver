# Kyver

Kyver is a production-ready starter foundation for an AI-powered software engineering platform.

## Tech Stack

- **Frontend:** Next.js 15, React, TypeScript, Tailwind CSS
- **Backend:** FastAPI (Python)
- **Database:** PostgreSQL
- **Authentication:** Better Auth (Next.js integration scaffold)
- **Infrastructure:** Docker, Docker Compose, GitHub Actions CI

## Architecture Highlights

- Modular frontend and backend structure for enterprise growth
- Versioned REST APIs (`/api/v1/...`)
- Frontend-to-backend API communication (`system/capabilities`)
- Better Auth route scaffold in Next.js (`/api/auth/*`)
- Environment-driven configuration via `.env`
- Backend capability registry that maps to future AI-agent modules

## Planned AI Platform Capabilities (Foundation Included)

- Multi-agent AI architecture
- Codebase understanding & repository indexing
- Architecture visualization
- Requirement analysis
- Code generation & code review
- Documentation generation & test case generation
- Change impact analysis
- RAG-based knowledge retrieval
- MCP integration
- Plugin system
- REST APIs

## Folder Structure

```text
.
├── frontend/                   # Next.js 15 app
│   └── src/
│       ├── app/               # App Router, auth route handlers
│       ├── features/          # Feature modules
│       ├── lib/               # API/auth/config utilities
│       └── types/             # Shared frontend types
├── backend/                    # FastAPI service
│   ├── app/
│   │   ├── api/v1/endpoints   # REST endpoint modules
│   │   ├── core               # App configuration
│   │   ├── db                 # Database session/engine setup
│   │   ├── schemas            # Response schemas
│   │   └── services           # Domain services / capability registry
│   └── tests/                 # Backend tests
├── .github/workflows/ci.yml    # CI pipeline
└── docker-compose.yml          # Local full-stack orchestration
```

## Environment Setup

1. Copy environment template:

```bash
cp .env.example .env
```

2. Optionally tailor values for local/dev.

## Local Development

### Frontend

```bash
cd frontend
npm install
npm run dev
```

### Backend

```bash
cd backend
python -m venv .venv
source .venv/bin/activate
pip install --upgrade pip
pip install .[dev]
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

## Docker

Run full stack (frontend + backend + postgres):

```bash
docker compose up --build
```

## CI

GitHub Actions workflow (`.github/workflows/ci.yml`) runs:

- Frontend: `npm ci`, `npm run lint`, `npm run build`
- Backend: `pip install .[dev]`, `ruff check .`, `pytest`

## API Endpoints

- `GET /` — backend service status
- `GET /api/v1/health` — health check
- `GET /api/v1/system/capabilities` — platform capability map

