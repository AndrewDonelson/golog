test:
	overalls -project=github.com/AndrewDonelson/golog -covermode=count
	go tool cover -func=./overalls.coverprofile
