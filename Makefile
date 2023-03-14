dev:
	@go build && MASTOK_MODE=DEV ./mastok
resetdev:
	@go build && MASTOK_MODE=RESET_DEV ./mastok && MASTOK_MODE=DEV ./mastok
buildfront:
	@go build && MASTOK_MODE=BUILD_FRONT ./mastok
deps:
	@go get -t -u ./... && go mod tidy
# see why "-p 1" here https://github.com/golang/go/issues/46959 + limiting the number
# of processes connected to the test DB
test:
	@clear && MASTOK_MODE=TEST MASTOK_PROJECT_ROOT=`pwd` GIN_MODE=release go test -p 1  ./...
testv:
	@clear && MASTOK_MODE=TEST MASTOK_PROJECT_ROOT=`pwd` GIN_MODE=release go test -p 1 -v ./...
testj:
	@clear && MASTOK_MODE=TEST MASTOK_PROJECT_ROOT=`pwd` GIN_MODE=release go test -p 1 -json ./...