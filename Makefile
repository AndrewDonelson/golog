test:
	overalls -project=github.com/vitalyisaev2/go-logging -covermode=count
	go tool cover -func=./overalls.coverprofile
