# Go API Template

A reusable REST API template written in Go, with PostgreSQL, JWT authentication, user management, password reset by email, audit logs and translated messages (pt-BR and en).

It is a port of `ts-node-api-template` (Node.js + TypeScript + Express + Prisma).

---

## 🧱 Tech Stack

- **Go** 1.26
- **chi** for routing and middlewares (`github.com/go-chi/chi/v5`)
- **pgx** as the PostgreSQL driver and connection pool (`github.com/jackc/pgx/v5`)
- **goose** for SQL migrations (`github.com/pressly/goose/v3`)
- **JWT** access and refresh tokens (`github.com/golang-jwt/jwt/v5`)
- **bcrypt** for password hashing (`golang.org/x/crypto/bcrypt`)
- **CORS** (`github.com/go-chi/cors`)
- **godotenv** to load the `.env` file
- Standard library for everything else: email (`net/smtp`), templates (`html/template`), translations (`embed`) and tests (`testing`)

---

## 📦 Requirements

- **Go** ≥ 1.26
- **PostgreSQL** running locally or in Docker

---

## 🚀 Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/PietroTelino/go-api-template.git
cd go-api-template
```

### 2. Download the dependencies

```bash
go mod download
```

### 3. Create the `.env` file

Copy `.env.example` to `.env` and fill in the values.

| Variable | Required | Default | Description |
|---|---|---|---|
| `PORT` | No | `3333` | HTTP port |
| `DB_HOST`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` | Yes | | PostgreSQL connection |
| `DB_PORT` | No | `5432` | PostgreSQL port |
| `JWT_SECRET` | Yes | | Secret used to sign access tokens |
| `JWT_EXPIRES_IN_SECONDS` | No | `900` | Access token lifetime (15 minutes) |
| `JWT_REFRESH_SECRET` | Yes | | Secret used to sign refresh tokens (use a different value) |
| `JWT_REFRESH_EXPIRES_IN_SECONDS` | No | `604800` | Refresh token lifetime (7 days) |
| `FRONTEND_URL` | Yes | | Allowed CORS origin and base URL of the password reset link |
| `EMAIL_ENABLED` | No | `false` | When `false`, emails are printed to the terminal instead of being sent |
| `EMAIL_HOST` | When email is enabled | | SMTP server |
| `EMAIL_PORT` | No | `587` | SMTP port |
| `EMAIL_SECURE` | No | `false` | `true` for implicit TLS (port 465); `false` uses STARTTLS when the server offers it |
| `EMAIL_USER`, `EMAIL_PASSWORD` | No | | SMTP credentials (authentication is skipped when the user is empty) |
| `EMAIL_FROM` | No | `no-reply@localhost` | Sender, e.g. `Go API <no-reply@example.com>` |
| `PASSWORD_RESET_EXPIRES_MS` | No | `3600000` | Password reset link lifetime (1 hour) |
| `GOD_EMAIL`, `GOD_PASSWORD` | Only for the seed | | Credentials of the first `god` user |

The API stops on startup when a required variable is missing.

---

## 🗄️ Database

### 1. Create the database

```bash
createdb <DB_NAME>
```

### 2. Apply the migrations

```bash
go run ./cmd/migrate up
```

Other commands:

```bash
go run ./cmd/migrate status   # list applied and pending migrations
go run ./cmd/migrate down     # revert the last migration (drops its table and data)
```

The migrations live in `migrations/` and are embedded in the binary.

### 3. Create the GOD user

```bash
go run ./cmd/seed
```

It uses `GOD_EMAIL` and `GOD_PASSWORD` and does nothing if the user already exists.

---

## 🏃 Running the Application

### Development

```bash
go run ./cmd/api
```

The server runs on `http://localhost:<PORT>`. `GET /health` returns `ok`.

### Build

```bash
go build -o bin/api ./cmd/api
```

The translations are embedded in the binary, so it only needs the environment variables to run. The same applies to `./cmd/migrate`, which embeds the SQL files.

---

## 🧪 Tests

```bash
go test ./...
```

The tests do not need a database. They cover the validation rules, the password generator, translations, JWT, the authentication middlewares, the email format and the audit filters.

---

## 📁 Project Structure

```txt
.
├── cmd
│   ├── api/            # HTTP server
│   ├── migrate/        # migration runner (up, down, status)
│   └── seed/           # creates the GOD user
│
├── internal
│   ├── auth/           # JWT, authentication and role middlewares
│   ├── config/         # environment variables
│   ├── database/       # PostgreSQL connection pool
│   ├── i18n/           # translations (pt-BR and en) and language middleware
│   ├── notifications/  # email sending
│   ├── response/       # JSON responses
│   ├── server/         # router and dependency wiring
│   ├── validation/     # shared validators
│   └── modules
│       ├── audit/           # audit log (write and read)
│       ├── authentication/  # login, refresh and logout
│       ├── passwordreset/   # password reset by email
│       ├── sessions/        # refresh tokens and active sessions
│       └── users/           # registration, profile and administration
│
├── migrations/         # SQL migrations (goose)
├── requests/           # .http files to test the routes
├── .env.example
└── go.mod
```

Each module follows the same layers: `handler` (HTTP) → `service` (business rules) → `repository` (SQL).

---

## 🔌 Routes

All routes are under `/api`.

| Method | Route | Access |
|---|---|---|
| `POST` | `/users` | Public: self registration |
| `POST` | `/auth/login` | Public |
| `POST` | `/auth/refresh` | Public |
| `POST` | `/auth/logout` | Authenticated |
| `POST` | `/auth/logout-all` | Authenticated |
| `POST` | `/password-reset/request` | Public |
| `POST` | `/password-reset/confirm` | Public |
| `PATCH` | `/users/me/preferences` | Authenticated |
| `PATCH` | `/users/me/password` | Authenticated |
| `DELETE` | `/users/me/delete` | Authenticated |
| `GET` | `/sessions/me` | Authenticated |
| `DELETE` | `/sessions/me/{sessionId}` | Authenticated |
| `GET` | `/audit/me` | Authenticated |
| `GET` | `/users` | Administrator |
| `GET` | `/users/deleted-users` | Administrator |
| `GET` | `/users/{id}` | Administrator |
| `POST` | `/users/create` | Administrator |
| `PUT` | `/users/{id}` | Administrator |
| `DELETE` | `/users/{id}` | Administrator (soft delete) |
| `PATCH` | `/users/{id}/restore` | Administrator |
| `PATCH` | `/users/{id}/inactivate` | Administrator |
| `PATCH` | `/users/{id}/reactivate` | Administrator |
| `POST` | `/users/{id}/reset-password` | Administrator |
| `GET` | `/audit` | Administrator |
| `GET` | `/audit/{id}` | Administrator |
| `DELETE` | `/users/{id}/hard` | God only (permanent delete) |

"Administrator" means the `administrator` and `god` roles. The `requests/` folder has ready-to-run examples for the REST Client extension of VS Code.

---

## 🔑 Authentication and Roles

- Login returns an **access token** (short-lived) and a **refresh token** (long-lived). Send the access token as `Authorization: Bearer <token>`.
- Each refresh token can be used only once: `/auth/refresh` revokes it and returns a new pair.
- Changing or resetting the password, inactivating or deleting a user revokes all of their sessions.
- An access token stays valid until it expires, even after logout, because it is not checked against the database.

There are three roles:

| Role | Can |
|---|---|
| `user` | Manage their own account |
| `administrator` | Manage users with the `user` role |
| `god` | Everything an administrator can, plus create administrators and delete users permanently |

A `god` user can only be created by the seed, and the admin routes cannot modify, inactivate, delete or reset the password of a `god` user.

---

## 🌐 Translations

Error and success messages are translated according to the `Accept-Language` header (`pt-BR` or `en`, defaulting to `pt-BR`). The files are in `internal/i18n/locales/`, and a test checks that both languages have the same keys.

---

## 🧹 Useful Commands

```bash
go run ./cmd/api              # run the API
go run ./cmd/migrate up       # apply migrations
go run ./cmd/migrate status   # migration status
go run ./cmd/seed             # create the GOD user
go test ./...                 # run the tests
go vet ./...                  # static analysis
go build -o bin/api ./cmd/api # build the binary
```

---

## 📜 License

This project is licensed under the **ISC License**.
Feel free to use and modify it for your own applications.
