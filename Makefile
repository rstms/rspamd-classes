# rspamd-classes  makefile

fmt: go.sum
	fix go fmt . ./...

build: fmt
	fix go build

test: testdata
	fix -- go test . ./...

debug: testdata
	fix -- go test . ./... -v --run $(test)

release: build test
	bump && gh release create v$$(cat VERSION) --notes "$$(cat VERSION)"


testdata:
	mkdir testdata

clean:
	go clean
	rm -f testdata/*.json

sterile: clean
	go clean -r -cache -modcache
	rm -f go.mod go.sum

go.sum: go.mod
	go mod tidy

go.mod:
	go mod init

