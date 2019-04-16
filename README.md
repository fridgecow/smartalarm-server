# Smart Alarm Server
HTTP Server for [Smart Alarm](https://www.github.com/fridgecow/smartalarm), a standalone sleep tracker for Wear OS.

## Installation

You will need an SMTP email server and a mySQL database.

### Download a Binary

There is at least one release binary [under the Releases tab](https://github.com/fridgecow/smartalarm-server/releases). It may be out of date. If you use this, skip to "Setup Environment"

### Build from Source

The server is written in Go. Once your go environment is setup, just run `go get github.com/fridgecow/smartalarm-server` to install `smartalarm-server` in `$GOPATH/bin`.

### Setup Environment

3. Run `db_structure.sql` against your mySQL DB to set up the required table.
4. Set up environment variables, listed below. E.g `export SMARTALARM_DBPASS=mydatabasepass`
6. Copy the `templates` directory in the working directory for the server wherever it's run. You must also create a `log` directory for the server to run.

The server runs on port 6662 by default. If localhost:6662/ping returns "☑", the server is up and running.

To use your server with Smart Alarm on your watch, you can edit the hostname in Settings. Since Wear OS's keyboard doesn't feature ":" by default, ";" will be replaced with ":".

## Environment Variables

These must be set or you will encounter errors. It may be wise to wrap the smartalarm-server binary in a bash script which sets these environment variables, or use some other deployment technique.

| Name                  | Usage                                          |
|-----------------------|------------------------------------------------|
|`SMARTALARM_DBHOST`    | Host for database connection                   |
|`SMARTALARM_DBNAME`    | Name of database to connect to                 |
|`SMARTALARM_DBPASS`    | Password for database access                   |
|`SMARTALARM_DBUSER`    | Username for database access                   |
|`SMARTALARM_EMAILADDR` | Address for SMTP connection and "From:" header |
|`SMARTALARM_EMAILHOST` | Host for SMTP connection                       |
|`SMARTALARM_EMAILPASS` | Password for email SMTP connection             |
|`SMARTALARM_EMAILREPLY`| Address for "ReplyTo:" header                  |

## API Description

- `GET /ping`: returns "☑", does not log. Useful to check if the server is up, although it does not perform any DB or SMTP connectivity checks.
- `GET /v1/add/[email]`: begins subscription process. Sends confirmation email to "email".
- `GET /v1/confirm/[email]/[token]`: confirms email address if "token" is correct for "email".
- `GET /v1/unsub/[email]/[token]`: unsubscribes email address if "token" is correct for "email".
- `POST /v1/csv/[email] csv=[csv_data] tz=[timezone]`: exports "csv_data" to "email", performing summaries and producing graphs. "timezone" should be the geographic timezone name, e.g "Europe/London", and this lets the server parse unix timestamps correctly.

## Possible Further Features

- Support for confirmation via email reply.
