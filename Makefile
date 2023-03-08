dev:
	@go build && MASTOK_MODE=DEV ./mastok
front:
	@go build && MASTOK_MODE=BUILD_FRONT ./mastok
deps:
	@go get -t -u ./... && go mod tidy
# see why "-p 1" here https://github.com/golang/go/issues/46959 + limiting the number
# of processes connected to the test DB
test:
	@MASTOK_MODE=TEST MASTOK_PROJECT_ROOT=`pwd` GIN_MODE=release go test -p 1  ./...
testv:
	@MASTOK_MODE=TEST MASTOK_PROJECT_ROOT=`pwd` GIN_MODE=release go test -p 1 -v ./...
testj:
	@MASTOK_MODE=TEST MASTOK_PROJECT_ROOT=`pwd` GIN_MODE=release go test -p 1 -json ./...