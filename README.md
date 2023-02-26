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

## Environment variables

The following environment variables are available:

- `MASTOK_ENV=DEV`:
    - load `.env` file to provide with a convenient way to define (other than `MASTOK_ENV`) environment variables
    - triggers automatic js processing (thanks to [esbuild](https://esbuild.github.io/))
- `MASTOK_LOGIN` and `MASTOK_PASSWORD` (both defaults to `admin`) to define login/password for HTTP basic authentication

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
