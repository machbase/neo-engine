all:
	GO111MODULE=on CGO_ENABLED=1 go build -o ./tmp/machgo ./main/machgo/main.go
	