# rspamd-class  makefile

fmt:
	fix go fmt . ./...

build: fmt
	fix go build

test:
	fix -- go test . ./...

release:
	bump
	gh release create v$(cat VERSION) --notes "v$(cat VERSION)"
