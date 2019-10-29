phony: test bench build run

test:
	go test -v .

bench:
	go test -bench .

build:
	go build examples/basic/main.go
	go build examples/http/main.go

run:
	./examples/basic/basic
	./examples/http/http