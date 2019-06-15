
compile-wasm:
	@ GOOS=js GOARCH=wasm go build -o main.wasm ./src/client/web/main

run-server: compile-wasm
	@ go run ./src/server/main/server.go

gotest:
	@ go test ./src/client ./src/ecs ./src/server

govet:
	@ go vet ./src/client ./src/ecs ./src/server ./src/game ./src/math
