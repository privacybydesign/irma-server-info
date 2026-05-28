# irma-server-info

Tiny HTTP service that records minimal info (email, version) reported by IRMA servers.

## Endpoint

`POST /` — body `{"email": "...", "version": "..."}`, header `User-Agent: irmaserver`. Duplicate `(email, version)` pairs are silently ignored.

## Build

Requires Go (version in [`go.mod`](go.mod)).

```sh
go build -o irma-server-info .
```

CI produces a static linux/amd64 binary; see [`.github/workflows/build.yml`](.github/workflows/build.yml).

## Run

```sh
./irma-server-info --config conf.yaml
```

If `conf.yaml` is missing the binary prints an example config and exits.

## Database

MySQL. Apply [`db.sql`](db.sql) once to create the database, user, and `servers` table. The default credentials in `db.sql` and `conf.yaml` are for local development only — replace them in production.
