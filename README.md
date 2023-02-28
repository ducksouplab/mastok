# Mastok

## Commands

Build and run

```
go build && MASTOK_ENV=DEV ./mastok
# or script alias
./dev
```

Run all tests
```
MASTOK_ENV=TEST MASTOK_PROJECT_ROOT=`pwd` GIN_MODE=release go test ./...
# or script alias
./testall
```

Build JS front files to `front/static` (and don't run mastok server)
```
go build && MASTOK_ENV=BUILD_FRONT ./mastok
```

## Environment variables

The following environment variables, regarding Mastok's own configuration:

- `MASTOK_ENV=DEV`:
    - load `.env` file to provide with a convenient way to define (other than `MASTOK_ENV`) environment variables
    - triggers automatic JS processing (thanks to [esbuild](https://esbuild.github.io/))
- `MASTOK_WEB_PORT` (defaults to `8190`) to set the port Mastok listens to
- `MASTOK_ORIGIN` (defaults to `http://localhost:8190`) to set what origin is trusted for WebSocket communication. If Mastok is running on port 8190 on localhost, but is served (thanks to a proxy) and reachable at https://mymastok.com, the valid `MASTOK_ORIGIN` value is `https://mymastok.com`
- `MASTOK_WEB_PREFIX` (defaults to `/`) if Mastok is served under a prefix path
- `MASTOK_LOGIN` and `MASTOK_PASSWORD` (both defaults to `admin`) to define login/password for HTTP basic authentication

And regarding connection to other services (no default values are provided):

- `MASTOK_DATABASE_URL` (like `postgres://ps_user:pg_password@localhost/mastok`) to connect to the database 
- `MASTOK_OTREE_URL` (like `http://localhost:8180/`) to reach oTree
- `MASTOK_OTREE_REST_KEY` to connect to oTree API

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

## Credits

This projects is in particular built upon [Gin](https://gin-gonic.com/), [GORM](https://gorm.io/), [esbuild](https://esbuild.github.io/).