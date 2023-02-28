dev:
	@go build && MASTOK_ENV=DEV ./mastok
front:
	@go build && MASTOK_ENV=BUILD_FRONT ./mastok
deps:
	@go get -t -u ./... && go mod tidy
test:
	@MASTOK_ENV=TEST MASTOK_PROJECT_ROOT=`pwd` GIN_MODE=release go test -v ./...
testv:
	@MASTOK_ENV=TEST MASTOK_PROJECT_ROOT=`pwd` GIN_MODE=release go test -v ./...