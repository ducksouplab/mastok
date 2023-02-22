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
MASTOK_PROJECT_ROOT=`pwd` GIN_MODE=release go test ./...
# or script alias
./test
```

## Environment variables

The following environment variables are available:

- `MASTOK_ENV=DEV`:
    - load `.env` file to provide with a convenient way to define (other than `MASTOK_ENV`) environment variables
    - triggers automatic js processing (thanks to [esbuild](https://esbuild.github.io/))
- `MASTOK_LOGIN` and `MASTOK_PASSWORD` (both defaults to `admin`) to define login/password for HTTP basic authentication