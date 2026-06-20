test:
	go test -v ./...

build:
	go build -o dist/encrypt-o-matic.exe ./cmd/encrypt-o-matic

server: build
	./dist/encrypt-o-matic.exe server --port 8080

clean:
	rm -rf dist build .encryptomatic demo.exe empty.exe large.exe testdir

.PHONY: test build server clean
