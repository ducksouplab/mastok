dev:
	@go build && MASTOK_MODE=DEV ./mastok
front:
	@go build && MASTOK_MODE=BUILD_FRONT ./mastok
deps:
	@go get -t -u ./... && go mod tidy
test:
	@MASTOK_MODE=TEST MASTOK_PROJECT_ROOT=`pwd` GIN_MODE=release go test ./...
testv:
	@MASTOK_MODE=TEST MASTOK_PROJECT_ROOT=`pwd` GIN_MODE=release go test -v ./...
testj:
	@MASTOK_MODE=TEST MASTOK_PROJECT_ROOT=`pwd` GIN_MODE=release go test -json ./...