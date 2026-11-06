# DB Lookout

DB Lookout is a self-hosted catalog browser for PostgreSQL and MySQL/MariaDB. It reads schema metadata and shows tables, columns, primary keys, and foreign keys in a small web UI. Database passwords stay on the server. The browser only receives connection names and catalog structure.

This repository is an early slice. It introspects live databases. It does not store descriptions, sign users in, or keep a history of schema changes.

## What you can do today

- Register one or more database servers in a JSON file.
- Give each server several databases and schemas.
- Browse a live catalog from PostgreSQL, MySQL, or MariaDB.
- Search the loaded tables, columns, and relations in the browser.
- Run the API and the UI together with Docker Compose or Podman Compose.

The account you configure should be read-only. The queries use `information_schema` and do not read table rows.

## Requirements

- A PostgreSQL, MySQL, or MariaDB server the app can reach.
- For Compose: Docker Compose v2 or Podman with a Compose provider.
- For local development: Go 1.27, Node.js 22, and npm.

## Quick start

From the repository root:

```bash
cp config/databases.example.json config/databases.json
cp .env.example .env
```

Edit `config/databases.json` with your host, port, database name, and schema list. Edit `.env` with the username and password named by that file. Then start both services:

```bash
podman compose up --build
```

Docker Compose uses the same file:

```bash
docker compose up --build
```

Open the UI at <http://localhost:5173>. The API listens on <http://localhost:8080>.

`config/databases.json` and `.env` are gitignored. Do not commit them.

## Configuration

`config/databases.json` is a list of connections. Each connection has one engine and one or more databases.

```json
{
  "connections": [
    {
      "id": "analytics-mysql",
      "name": "Analytics",
      "engine": "mysql",
      "host": "localhost",
      "port": 3306,
      "username_env": "DBLIB_ANALYTICS_USERNAME",
      "password_env": "DBLIB_ANALYTICS_PASSWORD",
      "sslmode": "disable",
      "databases": [
        {
          "id": "analytics-development",
          "name": "Analytics Development",
          "environment": "development",
          "database": "analytics_dev",
          "schemas": ["analytics_dev"]
        }
      ]
    }
  ]
}
```

| Field | Meaning |
| --- | --- |
| `id` | Stable id. Database ids must be unique across the whole file. |
| `name` | Label shown in the UI. |
| `engine` | `postgres`, `postgresql`, `mysql`, or `mariadb`. |
| `host`, `port` | Server address. PostgreSQL defaults to `5432`. MySQL and MariaDB default to `3306`. |
| `username_env`, `password_env` | Names of environment variables that hold the username and password. These fields are not the credentials themselves. |
| `sslmode` | `disable` turns TLS off. Any other value enables TLS for MySQL and is passed through for PostgreSQL. |
| `database` | Database name used when opening the connection. |
| `schemas` | Schemas to introspect. On MySQL and MariaDB, a schema is the database name. |
| `environment` | Label such as `development` or `production`. It is shown in the UI and is not sent to the database. |

The example file at `config/databases.example.json` contains one PostgreSQL connection and one MySQL connection.

### Credentials

Put the real username and password in `.env`, using the names from `username_env` and `password_env`:

```bash
DBLIB_RISK_USERNAME=dbdocs_reader
DBLIB_RISK_PASSWORD=replace-me
DBLIB_ANALYTICS_USERNAME=dbdocs_reader
DBLIB_ANALYTICS_PASSWORD=replace-me
```

Compose loads that file into the API container. The JSON file is mounted read-only at `/etc/db-library/databases.json`.

To use a registry file outside the repository, set `DBLIB_DATABASES_FILE` to its path before running Compose. If that variable is unset, Compose uses `./config/databases.json`.

`DBLIB_CONFIG` tells the API process where to read the registry inside the container. Compose sets it to `/etc/db-library/databases.json`. For a local process, leave it unset to use `config/databases.json` from the current working directory, or export `DBLIB_CONFIG` yourself.

## Local development

Start the API from the repository root. [Air](https://github.com/air-verse/air) rebuilds it when Go files change:

```bash
air -c .air.toml
```

In a second terminal, start the UI:

```bash
cd web
npm ci
npm run dev
```

Vite serves <http://localhost:5173> and proxies `/api` to <http://localhost:8080>. To point that proxy somewhere else, set `API_PROXY_TARGET`.

Run the tests from the Go module:

```bash
cd backend
go test ./...
```

To serve the built UI from the API process instead of Vite, build both parts and start the binary from the repository root:

```bash
cd web
npm ci
npm run build
cd ..
go build -C backend -o ./tmp/db-library ./cmd/db-library
./tmp/db-library
```

The API looks for `config/databases.json` and `web/dist` in its current working directory. Start it from the repository root. On Windows the binary is `tmp\db-library.exe`.

## API

Both routes accept `GET` only. Other methods return `405`.

`GET /api/databases` returns the connections and databases from the registry. The response includes ids, names, engine, environment, and schema names. It does not include hosts, usernames, or passwords.

`GET /api/databases/{id}/catalog` connects to the selected database and returns its catalog:

```json
{
  "name": "Analytics Development",
  "tables": [
    {
      "schema": "analytics_dev",
      "name": "events",
      "description": "",
      "columns": [
        {
          "name": "id",
          "type": "bigint",
          "nullable": false,
          "primary_key": true,
          "description": ""
        }
      ],
      "relations": [
        {
          "from_column": "user_id",
          "to_table": "users",
          "to_column": "id"
        }
      ]
    }
  ]
}
```

`description` is present on tables and columns and is empty. The app does not read or store comments yet. A database the server cannot reach returns `502` with `{"error":"unable to load database catalog"}`. An unknown id returns `404`.

## Layout

```text
compose.yaml          API and web services
backend/              Go module, API, and backend Dockerfile
web/                  Vue UI and frontend Dockerfile
config/               Example registry; your real file stays untracked
```

The API image is built from `backend/Dockerfile`. The web image is built from `web/Dockerfile` and runs the Vite dev server. Compose publishes the UI on port `5173` unless `APP_PORT` is set.

## Security

The API container runs as a non-root user, with a read-only root filesystem, all capabilities dropped, and `no-new-privileges`. The registry is mounted read-only. Keep the database account limited to metadata and the schemas you list. Set `sslmode` to a value other than `disable` when the server requires TLS.

## Not included yet

- Saved column descriptions, ownership, or sensitivity labels.
- Authentication and authorization.
- A cache, so every catalog view queries the database again.
- Schema history or a static documentation export.
- Reading database comment text into `description`.

## License

No license file is included yet. Add one before publishing the repository.
