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

JS and CSS files are processed by [esbuild](https://esbuild.github.io/) according to the configuration in `front/build.go`

Once `mastok` binary is built, launch JS/CSS processing with:

```
MASTOK_MODE=BUILD_FRONT ./mastok
```

Processed files are saved in `front/static/assets/[version]/[js|css]/` where `version` is defined and can be changed in `front/config.yml`. To ease things, whenever `version` is updated, the previous command also scans files in `templates/` and updates JS and CSS references, for instance:

```html
<script src="{{ WebPrefix }}/assets/v1.1/js/join.js"></script>
# becomes
<script src="{{ WebPrefix }}/assets/v1.2/js/join.js"></script>
```

This is useful to avoid cache issues.

### Build assets

To sumup, whenever you update and want to release a new version of JS or CSS files, you should first bump the `version` property in `front/config.yml` (always increase it, for instance from `v1.9` to `v1.10`), and then run:

```
MASTOK_MODE=BUILD_FRONT ./mastok
```

### Adding a new asset

Create a new file in `front/src/js` or `front/src/css`, declare it in `EntryPoints` in `front/build.go` and include it in your templates:

```html
# new js file
<script src="{{ WebPrefix }}/assets/v1.1/js/new.js"></script>
# new css file
<link rel="stylesheet" href="{{ WebPrefix }}/assets/v1.1/css/new.css">
```

Then process it, still with:

```
MASTOK_MODE=BUILD_FRONT ./mastok
```

## DuckSoup Docker image

Build image:

```
docker build -f docker/Dockerfile.build -t mastok:latest .
```

If you are building from mac add the --platform linux/amd64 so it works in linux:
```
docker build -f docker/Dockerfile.build -t mastok:latest . --platform linux/amd64
```

As an aside, this image is published on Docker Hub as `ducksouplab/mastok`, let's tag it and push it:

```
docker tag mastok ducksouplab/mastok
docker push ducksouplab/mastok:latest
```

## Types

There are shared types in the otree package (representing oTree REST API in and outs) and in the models package (saved to DB), there share data thoses their names and format is chosen to be closer to their usage. There are conversion functions when needed.

## Join sequence

When a participant arrives on the campaign join page (Share URL), here is a typical sequence:

- the server sends the current campaign State to the JS client: it must be `Running` to continue 
- the JS client sends a `Land` message to share a fingerpring that acts as an identifier (then server will decide to accept, redirect or ban this participant for this particular session)
- if `Land` is accepted by server:
    - the server sends a `Consent` message
    - the participant is asked to agree with it
    - if yes, the JS client sends an `Agree` message to the server
- if the campaign relies on grouping participants:
    - the server sends a `Grouping` message
    - the participant is asked to select a group (for instance male or female)
    - the JS client sends a `Connect` message to the server
    - if there was no grouping, `Connect` is not needed
- now the participant has joined the waiting room and a `Joining` message update is sent from the server (or `Pending` if there are too many people in the joining pool)
- when the joining pool is full (ready), the client receives a `SessionStart` message from the server

## Busy state

The `Busy` state means the maximum number of concurrent sessions is currently reached and new ones are on hold. This state is checked and updated periodically.

New participants can't join, but they will be added to the pending pool.

When we switch from `Busy` to `Running` state (a session has finished):
- participants from the pending pool are used to fill the joining pool
- if there are enough people in the joining pool, a new session is created, possibly switching the state back to `Busy`

## Participant messages

The server may send the following messages to a participant: `State`, `Consent`, `Grouping`, `Pending`, `Joining`, `Instructions`, `SessionStart`, `Paused`, `Completed`, `Disconnect`.

A participant may send the following messages to the server: `Land`, `Agree`, `Connect`.

## Supervisor messages

The server may send the following messages to a supervisor: `State`, `JoiningSize`, `PendingSize`, `SessionStart`.

A supervisor may send the following messages to the server: `State` (to update it).

## Joining and fingerprinting

JS fingerprinting is used to identify unique users (even if it's not 100% reliable). It is both helpful for:

- #1 preventing the same participant to be involved several times (= in several sessions) of the same campaign. This applies only if `JoinOnce` is true for this campaign
- #2 helping a user reconnecting to a live (= currently running) session if, by mistake, they closed their tab and come back using mastok's slug (that will then redirect on oTree). This is mastok default's behaviour and can be disabled by setting the env variable: `MASTOK_DISABLE_LIVE_REDIRECT=true`

As a result for developers/testers (that open many tabs with the same fingerprint):

- Due to #1 and if `JoinOnce` is true, they will face un `Unavailable` page the second time they use the public slug

- If `JoinOnce` is false, they can open many tabs if the session is not started, and join oTree "separately" (as a different participant) from each tab. But due to #2, if the session is running, the will be redirected to oTree

## Credits

This projects is in particular built upon [Gin](https://gin-gonic.com/), [GORM](https://gorm.io/), [esbuild](https://esbuild.github.io/) and [gorilla](https://github.com/gorilla/websocket).