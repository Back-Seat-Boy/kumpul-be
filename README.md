# kumpul-be

`kumpul-be` is the backend service for Kumpul, a group event coordination app for recurring hangouts and sports sessions. It handles Google sign-in, event planning, option voting, RSVPs, payment tracking, WhatsApp message link generation, and image upload for payment proofs.

## Stack

- Go `1.26.1`
- Echo v4 for HTTP delivery
- PostgreSQL-compatible SQL database
- Redis for sessions and caching
- Cloudinary for image uploads
- Cobra for CLI commands
- Viper for configuration

## What the Service Does

- Google OAuth login with frontend redirect callback
- Session-based auth using `Authorization: Bearer <session_id>`
- Venue management
- Event creation and status updates
- Event option creation, listing, and deletion
- Voting on event options
- Participant join, leave, guest join, and removal
- Payment creation, claim, confirmation, and adjustment
- WhatsApp deep links for venue booking and payment nudges
- Image upload to Cloudinary

## Project Layout

```text
.
├── main.go
├── config.yml.example
├── docker-compose.yml
├── db/migrations/
├── internal/config/
├── internal/console/
├── internal/delivery/
├── internal/model/
├── internal/repository/
└── internal/usecase/
```

The code follows a layered structure:

- `internal/delivery` handles HTTP transport and response formatting
- `internal/usecase` contains business logic
- `internal/repository` manages persistence and cache access
- `internal/model` defines domain models, contracts, and shared errors
- `internal/console` wires the app and CLI commands together

## Requirements

- Go `1.26.1`
- Docker and Docker Compose, or local PostgreSQL and Redis instances
- Google OAuth credentials
- Cloudinary credentials

## Configuration

Copy the example config:

```bash
cp config.yml.example config.yml
```

Important settings in `config.yml`:

- `port`: HTTP port, default `8080`
- `database.host`, `database.database`, `database.username`, `database.password`
- `redis.cache_host`, `redis.lock_host`
- `google.client_id`, `google.client_secret`, `google.redirect_url`
- `cors.allowed_origins`
- `frontend_url`
- `cloudinary.cloud_name`, `cloudinary.api_key`, `cloudinary.api_secret`, `cloudinary.upload_folder`

Notes:

- The app reads `config.yml` from the project root.
- In non-production environments, the database DSN uses `sslmode=disable`.
- The sample Google callback URL is `http://localhost:8080/auth/google/callback/`.

## Local Development

Start dependencies:

```bash
docker-compose up -d
```

Run database migrations:

```bash
go run main.go migrate --direction up
```

Start the API server:

```bash
go run main.go server
```

The service will be available at `http://localhost:8080`.

## CLI Commands

Run the server:

```bash
go run main.go server
```

Run migrations:

```bash
go run main.go migrate --direction up
go run main.go migrate --direction down --step 1
```

Build the binary:

```bash
go build -o bin/kumpul main.go
```

Run tests:

```bash
go test ./...
```

## API Overview

Public endpoints:

- `GET /ping/`
- `GET /auth/google/login/`
- `GET /auth/google/callback/`
- `GET /events/:token/`
- `GET /events/:token/options/`
- `GET /events/:token/options/with-voters/`
- `GET /events/:token/participants/`

Protected endpoints live under `/api` and require:

```http
Authorization: Bearer <session_id>
```

Examples:

- `GET /api/users/me/`
- `GET /api/venues/`
- `POST /api/events/`
- `POST /api/events/:event_id/votes/`
- `POST /api/events/:event_id/participants/guest/`
- `POST /api/events/:event_id/payment/`
- `POST /api/uploads/image/`

There is a fuller endpoint reference in [`API_SPECIFICATION.md`](/Users/bagasp/VsCodeProjects/kumpul-be/API_SPECIFICATION.md).

## Authentication Flow

1. Frontend requests `GET /auth/google/login/`
2. Backend returns a Google OAuth URL
3. User authenticates with Google
4. Google redirects to `/auth/google/callback/`
5. Backend creates or updates the user, creates a session, then redirects to the frontend callback route with session and user query params
6. Frontend stores the returned `session_id` and sends it as a Bearer token on protected requests

## Uploads

The upload endpoint accepts a base64-encoded image payload and stores it in the configured Cloudinary folder. The response includes the hosted URL and Cloudinary public ID.

## Useful Files

- [`main.go`](/Users/bagasp/VsCodeProjects/kumpul-be/main.go)
- [`internal/console/server.go`](/Users/bagasp/VsCodeProjects/kumpul-be/internal/console/server.go)
- [`internal/delivery/routes.go`](/Users/bagasp/VsCodeProjects/kumpul-be/internal/delivery/routes.go)
- [`config.yml.example`](/Users/bagasp/VsCodeProjects/kumpul-be/config.yml.example)
- [`docker-compose.yml`](/Users/bagasp/VsCodeProjects/kumpul-be/docker-compose.yml)

## Current Notes

- The codebase uses PostgreSQL terminology in config and migrations, while the connection bootstrap currently uses `go-connect`'s CockroachDB helper with a PostgreSQL DSN.
- The repository already contains a richer API specification than the old README, so this README is intended as a practical entry point rather than a full contract document.
