phony: test build run

test:
	go test -bench .
	# overalls -project=github.com/AndrewDonelson/golog -covermode=count
	# go tool cover -func=./overalls.coverprofile

build:
	cd example && go build

run:
	./example/example