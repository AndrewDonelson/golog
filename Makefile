phony: build

test:
	overalls -project=github.com/AndrewDonelson/golog -covermode=count
	go tool cover -func=./overalls.coverprofile

build:
	cd example && go build
