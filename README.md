# Omnibus Backend

Backend for a bus terminal management system. It manages terminals, bays, cities, users and real-time notifications to passengers.

## Entry point

```
cmd/api/main.go
```

Startup follows this flow: load config → connect to PostgreSQL → initialize the dependency container (`app.New`) → register routes → start the server.

## Main technologies and libraries

| Library | Use |
|---|---|
| `github.com/labstack/echo/v4` | HTTP framework |
| `gorm.io/gorm` + `gorm.io/driver/postgres` | ORM and PostgreSQL driver |
| `github.com/golang-jwt/jwt/v5` | JWT authentication |
| `github.com/philippseith/signalr` | Real-time communication (SignalR) |
| `golang.org/x/crypto` | Password hashing (bcrypt) |
| `github.com/google/uuid` | UUID generation |
| `github.com/joho/godotenv` | Loading environment variables from `.env` |

## Getting started

### Prerequisites

- Go 1.25+
- PostgreSQL running locally (or configurable via `DATABASE_URL`)
- `make` installed

### Environment variables

Create a `.env` file at the root (or export the variables to the environment):

```env
DATABASE_URL=postgres://postgres:1234@localhost:5432/omnibus-terminal
JWT_SECRET=your-jwt-secret
PASSWORD_RESET_JWT_SECRET=your-reset-secret

LISTEN_ADDR=:4989

SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=email@gmail.com
SMTP_PASSWORD=app-password
SMTP_FROM=email@gmail.com

EXTERNAL_TERMINAL_UPSTREAM_URL=http://localhost:4990
REALTIME_URL=http://localhost:4988/realtime
REALTIME_API_KEY=realtime-key
CAMERA_NOTIFICATION_API_KEY=camera-key

FRONT_END_BASE_LINK=http://localhost:4200/
```

### Run the server

```bash
go run ./cmd/api
```

## Migrations

The schema is managed with the `cmd/migrate` command, which uses the same `DB_*` variables as the API and records what has been applied in the `omnibus_schema_migrations` table. It does not require `psql`.

```bash
go run ./cmd/migrate up        # or: make migrate-up
go run ./cmd/migrate status    # which migrations are applied
go run ./cmd/migrate down      # reverts the last one
go run ./cmd/migrate seed      # sample data (migrations/seed.sql)
go run ./cmd/migrate seed-ar   # cities and terminals of Argentina (or: make migrate-seed-ar)
go run ./cmd/migrate baseline  # marks all of them as applied without running them
```

- **Empty database:** `up` runs `migrations/schema.sql` (full current schema + roles) and marks every migration as applied.
- **Existing database:** `up` applies only the pending `NNN_name.up.sql` migrations, in order and each one in its own transaction.
- **Database created before this command** (by hand with `psql`): run `baseline` exactly once.
- **Cities and terminals of Argentina:** `seed-ar` reads `migrations/data/cities.json` and `migrations/data/terminals.json`. It also calls `GET {EXTERNAL_TERMINAL_UPSTREAM_URL}/terminal/` on backend-terminales and fills in each terminal's `external_terminal_id` by looking up its `external_name` (ignoring case and accents). If backend-terminales does not respond, it fails without touching the database. When it finishes it reports how many `external_terminal_id`s it set and lists the terminals with no match. It can be run as many times as you like: it does not overwrite existing cities, only adds terminals to cities that do not have any yet, and does not replace an already assigned `external_terminal_id`. To add more, edit the JSON files and run it again. In Docker: `/app/migrate seed-ar`.

In production the Docker image includes the binary: `/app/migrate up` (for example as the *Pre-Deploy Command* on Railway).

When adding a new migration, create `NNN_name.up.sql` / `NNN_name.down.sql` and **reflect the change in `schema.sql` too**.

| Version | Description |
|---|---|
| `001` | Base tables: `city`, `bus_terminal`, `platform` |
| `002` | Auth and roles: `rol`, `users`, `user_refresh_tokens`, `user_terminal` + roles and permissions seed |
| `003` | CASCADE on the `bus_terminal → city` FK |
| `004` | `external_terminal_id` field on `bus_terminal` |
| `005` | Removal of the `dni` field from `users` |
| `006` | `awaited_trip` table |
| `007` | Removal of `permissions` and `rol_permissions` (RBAC simplified to roles) |
| `008` | Wait data in `awaited_trip` (ticket, terminal, `notified_at`) |
| `009` | `notifications` table (it existed in local databases but had no migration) |
| `010` | English names: `platform.anden` → `bay`, `rol` → `role`, `users.rol_id` → `role_id`; renames `anden` / `license_patent` keys in stored notifications |

## Architecture

The project follows a strict layered architecture:

```
Handler → Service → Repository → Database (PostgreSQL/GORM)
```

```
cmd/api/            Entrypoint
cmd/migrate/        Migration runner
internal/
├── app/            Dependency container (manual DI)
├── config/         Configuration loading (.env + defaults)
├── database/       PostgreSQL connection
├── errors/         Reusable domain errors
├── handler/        HTTP controllers (HTTP input/output)
├── mail/           Email service + HTML templates
├── middleware/     CORS, logging, JWT auth, roles, API key
├── models/         Domain structs and DTOs
├── realtime/       SignalR client (hub groups and methods)
├── repository/     Data access via GORM
├── roles/          Role constants (user, admin, super_admin)
├── router/         Route registration per profile
├── service/        Business logic
└── validators/     Input validation per domain
pkg/
└── realtime/       SignalR client wrapper
migrations/         Versioned SQL files
```

**Applied patterns:**
- Manual Dependency Injection in `app.New()`
- Repository Pattern to abstract data access
- Middleware chain: CORS → Logging → JWT Auth → Role check → Handler
- Short-lived access JWT + Refresh Token in an HTTP-only cookie

**Role system:**

| Role | Description |
|---|---|
| `user` | Registered passenger |
| `admin` | Administrator of a terminal |
| `super_admin` | Global system administrator |

## Endpoints

The full documentation (OpenAPI 3) is in [internal/docs/openapi.yaml](internal/docs/openapi.yaml) and is served by the server:

- `GET /docs`: Swagger UI to explore and try the endpoints.
- `GET /docs/openapi.yaml`: the specification (can be imported into Postman, Insomnia, etc.).

### Public / Auth

| Method | Path | Description |
|---|---|---|
| `GET` | `/` | Redirects to `/health` |
| `GET` | `/health` | Health check (`OK`) |
| `GET` | `/docs` | Swagger UI with the OpenAPI documentation |
| `GET` | `/docs/openapi.yaml` | OpenAPI specification |
| `GET` | `/bus_tickets/:ticket_string` | Look up a bus ticket by ticket string |
| `POST` | `/notify_passengers` | Notifies passengers of a bus arrival (API key, or `admin`/`super_admin` JWT for the manual notice; the admin only on bays of their terminals) |
| `POST` | `/notify_camera_error` | Notifies a camera error (API key) |
| `GET` | `/api/notifications` | Lists the user's notifications (JWT) |
| `POST` | `/api/auth/register` | Registers a new user |
| `POST` | `/api/auth/login` | Login, returns an access token |
| `POST` | `/api/auth/refresh` | Refreshes the access token |
| `POST` | `/api/auth/logout` | Logs out |
| `POST` | `/api/auth/forgot-password` | Requests a recovery email |
| `POST` | `/api/auth/validate-recovery-token` | Validates a recovery token |
| `POST` | `/api/auth/reset-password` | Resets the password |

### Authenticated user (`user`, `admin`, `super_admin`)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/users/me` | Gets own profile |
| `PUT` | `/api/users/me` | Updates own profile |
| `DELETE` | `/api/users/me` | Deletes own account |
| `GET` | `/api/users/terminals` | Lists the terminals |
| `POST` | `/api/buses/join` | Starts waiting for a trip (validates the bus ticket and returns the realtime `group_key`) |
| `GET` | `/api/buses/awaited` | Trip the user is waiting for |
| `DELETE` | `/api/buses/awaited` | Stops waiting for the trip |

### Admin (`admin`, `super_admin`)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/admin/cities` | Lists cities |
| `GET` | `/api/admin/cities/count` | Number of cities |
| `GET` | `/api/admin/cities/:postal_code` | Gets a city by postal code |
| `GET` | `/api/admin/platforms` | Lists bays |
| `GET` | `/api/admin/platforms/:code` | Gets a bay by code |
| `POST` | `/api/admin/platforms` | Creates a bay |
| `PUT` | `/api/admin/platforms/:code` | Updates a bay |
| `DELETE` | `/api/admin/platforms/:code` | Deletes a bay |
| `GET` | `/api/admin/users/by-email` | Finds a user by email |
| `POST` | `/api/admin/users/promote` | Promotes a user to admin |
| `POST` | `/api/admin/users/demote` | Demotes an admin to user |
| `GET` | `/api/admin/notification-types` | Lists the available notification types |
| `POST` | `/api/admin/notifications` | Sends a notification to passengers |
| `DELETE` | `/api/admin/notifications` | Deletes a notification |
| `POST` | `/api/admin/notify-bus-delay` | Notifies a bus delay |

### Super Admin (`super_admin`)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/super/terminals` | Lists all terminals |
| `GET` | `/api/super/terminals/:uuid` | Gets a terminal by UUID |
| `POST` | `/api/super/terminals` | Creates a terminal |
| `PUT` | `/api/super/terminals/:uuid` | Updates a terminal |
| `DELETE` | `/api/super/terminals/:uuid` | Deletes a terminal |
| `POST` | `/api/super/users/promote-super` | Promotes an admin to super_admin |
| `POST` | `/api/super/users/demote-super` | Demotes a super_admin to admin |
| `POST` | `/api/admin/cities` | Creates a city |
| `PUT` | `/api/admin/cities/:postal_code` | Updates a city |
| `DELETE` | `/api/admin/cities/:postal_code` | Deletes a city |
