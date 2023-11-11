run_server:
	go run ./cmd/dbouncer/main.go

run_client:
	go run ./cmd/go_client/main.go

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o ./bin/dbouncer ./cmd/dbouncer/main.go
	cp ./config.yml ./bin/config.yml