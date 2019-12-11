
phony: test bench build run

test:
	go test -v .

bench:
	go test -bench .

build:
	go build -o ./examples/basic/basic examples/basic/main.go
	go build -o ./examples/http/http examples/http/main.go

run:
	./examples/basic/basic
	./examples/http/http
