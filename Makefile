# rspamd-classes  makefile

fmt:
	fix go fmt 

build: fmt
	fix go build

test: testdata
	fix -- go test

debug: testdata
	fix -- go test -v --run $(test)

release:
	bump && gh release create v$$(cat VERSION) --notes "$$(cat VERSION)"


testdata:
	mkdir testdata

clean:
	go clean
	rm -f testdata/*.json

