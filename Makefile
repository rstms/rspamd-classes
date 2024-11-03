# rspamd-class  makefile

fmt:
	fix go fmt . ./...

build: fmt
	fix go build

test:
	fix -- go test . ./...
