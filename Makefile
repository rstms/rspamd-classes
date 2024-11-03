# rspamd-class  makefile

fmt:
	fix go fmt . ./...

build: fmt
	fix go build

test:
	fix -- go test . ./...

release:
	bump
	gh release create v$(shell cat VERSION) --notes "v$(shell cat VERSION)"
