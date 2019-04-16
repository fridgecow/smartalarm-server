# Smart Alarm Server
HTTP Server for [Smart Alarm](https://www.github.com/fridgecow/smartalarm), a standalone sleep tracker for Wear OS.

## Installation

The server is written in Go, so make sure your go environment is set up. You will also need an SMTP email server and a mySQL database.

1. `git clone`
2. `go get .` to fetch dependencies.
3. Run `db_structure.sql` against your mySQL DB to set up the required table.
4. Set up environment variables, listed below. E.g `export SMARTALARM_DBPASS=mydatabasepass`
5. `go run .` will run the server immediately. `go install .` or `go build .` will create a binary for installation elsewhere.
6. Copy the `templates` directory in the working directory for the server wherever it's run. You must also create a `log` directory for the server to run.

The server runs on port 6662 by default. If localhost:6662/ping returns "☑", the server is up and running.

## Environment Variables

These must be set or you will encounter errors.

| Name                 | Usage                              |
|----------------------|------------------------------------|
|`SMARTALARM_DBPASS`   | Password for database access       |
|`SMARTALARM_EMAILPASS`| Password for email SMTP connection |

## API Description

- `GET /ping`: returns "☑", does not log. Useful to check if the server is up, although it does not perform any DB or SMTP connectivity checks.
- `GET /v1/add/[email]`: begins subscription process. Sends confirmation email to "email".
- `GET /v1/confirm/[email]/[token]`: confirms email address if "token" is correct for "email".
- `GET /v1/unsub/[email]/[token]`: unsubscribes email address if "token" is correct for "email".
- `POST /v1/csv/[email] csv=[csv_data] tz=[timezone]`: exports "csv_data" to "email", performing summaries and producing graphs. "timezone" should be the geographic timezone name, e.g "Europe/London", and this lets the server parse unix timestamps correctly.

## Possible Further Features

- Support for confirmation via email reply.
