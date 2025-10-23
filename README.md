# Chat Application

Slack-like communication application with real-time messaging, workspaces, channels, and file attachments.

## Tech Stack

### Backend
- Go 1.22+
- Gin (HTTP router)
- WebSocket (gorilla/websocket)
- GORM + Gen (ORM & code generation)
- Atlas (declarative schema migration)
- PostgreSQL
- JWT authentication
- Wasabi S3-compatible storage

### Frontend
- React 19
- TypeScript
- Vite
- Mantine 8 (UI components)
- Tailwind CSS
- TanStack Router
- TanStack Query
- PWA support (vite-plugin-pwa)
- Vitest + Storybook

### Infrastructure
- Docker Compose
- Caddy (reverse proxy)
- VPS deployment ready

## Project Structure

```
chat/
├── backend/          # Go backend
│   ├── cmd/
│   │   └── server/  # Main application entry point
│   ├── internal/
│   │   ├── domain/         # Domain entities & repository interfaces
│   │   ├── usecase/        # Business logic (to be implemented)
│   │   ├── interface/
│   │   │   ├── http/       # HTTP handlers & routes
│   │   │   └── ws/         # WebSocket hub & connections
│   │   └── infrastructure/
│   │       ├── auth/       # JWT & password hashing
│   │       ├── config/     # Configuration management
│   │       ├── db/         # GORM models & connection
│   │       ├── logger/     # Zap logger setup
│   │       └── storage/    # Wasabi S3 client (to be implemented)
│   ├── schema/       # Atlas declarative schema (HCL)
│   └── atlas.hcl     # Atlas configuration
├── frontend/         # React frontend
│   ├── src/
│   │   ├── routes/   # TanStack Router routes
│   │   ├── features/ # Feature-based modules
│   │   ├── components/ # Reusable UI components
│   │   └── lib/      # API client, WS client, etc.
│   └── public/       # Static assets & PWA manifest
├── docker/           # Docker configurations
└── schema/           # Shared schema files

```

## Current Implementation Status

### Completed
- [x] Monorepo structure (pnpm workspaces + Turbo)
- [x] Backend Clean Architecture skeleton
- [x] Domain entities (User, Workspace, Channel, Message, etc.)
- [x] Repository interfaces
- [x] JWT authentication infrastructure
- [x] Password hashing (bcrypt)
- [x] Configuration management
- [x] Logging (zap)
- [x] HTTP middleware (CORS, auth, rate limiting)
- [x] WebSocket hub & connection management
- [x] GORM models for all entities
- [x] OpenAPI 3.1 specification
- [x] Atlas schema definition (PostgreSQL)

### In Progress
- [ ] Repository implementations with GORM
- [ ] Use case layer (business logic)
- [ ] HTTP handlers for all endpoints
- [ ] Wasabi S3 client implementation

### Planned
- [ ] Frontend initialization
- [ ] OpenAPI client generation
- [ ] TanStack Router & Query setup
- [ ] Chat UI components
- [ ] WebSocket integration
- [ ] PWA features
- [ ] Testing (Vitest)
- [ ] Storybook
- [ ] Docker & deployment setup

## Getting Started

### Prerequisites
- Go 1.22+
- Node.js 20+
- pnpm 10+
- PostgreSQL 15+
- Atlas CLI (for migrations)

### Backend Setup

1. Install dependencies:
```bash
cd backend
go mod download
```

2. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Run database migrations:
```bash
atlas migrate apply --env dev
```

4. Run the server:
```bash
go run cmd/server/main.go
```

### Frontend Setup

1. Install dependencies:
```bash
cd frontend
pnpm install
```

2. Run the development server:
```bash
pnpm dev
```

## API Documentation

The API is documented using OpenAPI 3.1. See [backend/internal/openapi/openapi.yaml](backend/internal/openapi/openapi.yaml) for the full specification.

### Key Endpoints

- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login
- `POST /api/auth/refresh` - Refresh access token
- `GET /api/workspaces` - List workspaces
- `GET /api/workspaces/{id}/channels` - List channels
- `GET /api/channels/{id}/messages` - List messages
- `POST /api/channels/{id}/messages` - Send message
- `GET /ws?workspaceId={id}` - WebSocket connection

## Database Schema

The database schema is managed by Atlas using declarative HCL files. See [backend/schema/schema.hcl](backend/schema/schema.hcl) for the complete schema definition.

### Main Tables
- `users` - User accounts
- `sessions` - JWT refresh tokens
- `workspaces` - Workspace containers
- `workspace_members` - Workspace membership & roles
- `channels` - Communication channels
- `channel_members` - Private channel membership
- `messages` - Chat messages (with thread support)
- `message_reactions` - Message reactions (emoji)
- `channel_read_states` - Unread message tracking
- `attachments` - File attachment metadata

## Development

### Running Tests
```bash
# Backend
cd backend
go test ./...

# Frontend
cd frontend
pnpm test
```

### Code Generation
```bash
# Generate OpenAPI types for frontend
cd frontend
pnpm run generate:api
```

## License

MIT
