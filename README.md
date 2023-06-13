# Mastok

## Commands

Build and run

```
go build && MASTOK_MODE=DEV ./mastok
# or script alias
make dev
```

Run all tests
```
MASTOK_MODE=TEST MASTOK_PROJECT_ROOT=`pwd` GIN_MODE=release go test ./...
# or script alias
make test
```

Build JS front files to `front/static` (and don't run mastok server)
```
go build && MASTOK_MODE=BUILD_FRONT ./mastok
# or script alias
make front
```

Update go modules dependencies
```
go get -t -u ./... && go mod tidy
# or script alias
make deps
```

## Environment variables

The following environment variables, regarding Mastok's own configuration:

- `MASTOK_MODE=DEV`:
    - load `.env` file to provide with a convenient way to define (other than `MASTOK_MODE`) environment variables
    - triggers automatic JS processing (thanks to [esbuild](https://esbuild.github.io/))
- `MASTOK_PORT` (defaults to `8190`) to set the port Mastok listens to
- `MASTOK_ORIGIN` (defaults to `http://localhost:8190`) to set what origin is trusted for WebSocket communication. If Mastok is running on port 8190 on localhost, but is served (thanks to a proxy) and reachable at https://mymastok.com, the valid `MASTOK_ORIGIN` value is `https://mymastok.com`
- `MASTOK_WEB_PREFIX` (defaults to nothing) if Mastok is served under a prefix path (for instance `/path` without trailing slash)
- `MASTOK_LOGIN` and `MASTOK_PASSWORD` (both defaults to `mastok`) to define login/password for HTTP basic authentication

And regarding connection to other services (no default values are provided):

- `MASTOK_DATABASE_URL` (like `postgres://ps_user:pg_password@localhost/mastok`) to connect to the database 
- `MASTOK_OTREE_PUBLIC_URL` (like `http://host.com/otree`) to reach oTree public pages
- `MASTOK_OTREE_API_URL` (like `http://localhost:8180`) to reach oTree REST API
- please note having two `MASTOK_OTREE_*_URL` may be useful if Mastok connects to the REST API in a different way than the clients (participants), but they should point to the same running oTree instance
- `MASTOK_OTREE_API_KEY` to authenticate to oTree API

## Front-end processing

JS and CSS files are processed by [esbuild](https://esbuild.github.io/) as defined in `front/build.go`.

This can be launche with:

```
MASTOK_MODE=BUILD_FRONT ./mastok
```

This step also scans files in `templates/` and updates JS and CSS includes with the version defined in `front/config.yml`, for instance:

```html
<script src="{{ WebPrefix }}/assets/v1.1/js/join.js"></script>
# changed to
<script src="{{ WebPrefix }}/assets/v1.2/js/join.js"></script>
```

This is useful in particular to avoid cache issues.

### Build assets

To sumup, whenever you update and want to release a new version of JS or CSS files, you should first bump the `version` property in `front/config.yml` (always increase it, for instance from `v1.9` to `v1.10`), and then:

```
MASTOK_MODE=BUILD_FRONT ./mastok
```

### Adding a new asset

Create a new file in `front/src/js` or `front/src/css`, declare it in `EntryPoints` in `front/build.go` and finally include it in your templates:

```html
# new js file
<script src="{{ WebPrefix }}/assets/v1.1/js/new.js"></script>
# new css file
<link rel="stylesheet" href="{{ WebPrefix }}/assets/v1.1/css/new.css">
```

## DuckSoup Docker image

Build image:

```
docker build -f docker/Dockerfile.build -t mastok:latest .
```

As an aside, this image is published on Docker Hub as `ducksouplab/mastok`, let's tag it and push it:

```
docker tag mastok ducksouplab/mastok
docker push ducksouplab/mastok:latest
```

## Types

There are shared types in the otree package (representing oTree REST API in and outs) and in the models package (saved to DB), there share data thoses their names and format is chosen to be closer to their usage. There are conversion functions when needed.

## Join sequence

When participant arrives on the campaign join page (Share URL), here is a typical sequence:

- the server returns current campaign State: it must be `Running` to continue 
- the client sends a `Land` message to share a fingerpring that acts as an identifier (then server will decide to accept, redirect or ban this participant for this particular session)
- if `Land` is accepted, the participant is asked to agree with the session rules
- if yes, the client sends a `Agree` message to the server
- if campaign relies on grouping participants, the participant is asked to select a group (for instance male or female), then the js client sends a `Connect` message to the server. If there is not grouping, `Connect` is not needed
- now the participant has joined the waiting room and a `Joining` message update is sent from the server (or `Pending` if there are too many people in the joining pool)
- when the joining pool is full (ready), the client receives a `SessionStart` message from the server

## Participant messages

The server may send the following messages to a participant: `State`, `Consent`, `Grouping`, `Pending`, `Joining`, `Instructions`, `SessionStart`, `Disconnect`.

A participant may send the following messages to the server: `Land`, `Agree`, `Connect`.

## Supervisor messages

The server may send the following messages to a supervisor: `State`, `JoiningSize`, `PendingSize`, `SessionStart`.

A supervisor may send the following messages to the server: `State` (to update it).

## Credits

This projects is in particular built upon [Gin](https://gin-gonic.com/), [GORM](https://gorm.io/), [esbuild](https://esbuild.github.io/) and [gorilla](https://github.com/gorilla/websocket) following this chat [example](https://github.com/gorilla/websocket/tree/master/examples/chat).