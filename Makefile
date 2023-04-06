dev:
	@go build && MASTOK_MODE=DEV ./mastok
newdev:
	@go build && MASTOK_MODE=RESET_DEV ./mastok && MASTOK_MODE=DEV ./mastok
buildfront:
	@go build && MASTOK_MODE=BUILD_FRONT ./mastok
deps:
	@go get -t -u ./... && go mod tidy
# see why "-p 1" here https://github.com/golang/go/issues/46959 + limiting the number
# of processes connected to the test DB
cleantest:
	@go clean -testcache
test:
	@clear && MASTOK_MODE=TEST MASTOK_PROJECT_ROOT=`pwd` go test -p 1  ./...
testv:
	@clear && MASTOK_MODE=TEST MASTOK_PROJECT_ROOT=`pwd` go test -p 1 -v ./...
testfmt:
	@clear && MASTOK_MODE=TEST MASTOK_PROJECT_ROOT=`pwd` go test -p 1 -json ./... 2>&1 | tee /tmp/gotest.log | gotestfmt -hide all
dockerbuild:
	@docker build -f docker/Dockerfile.build -t mastok:latest . && docker tag mastok ducksouplab/mastok
dockerpush:
	@docker push ducksouplab/mastok:latest
