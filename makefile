
compile-wasm:
	echo "Builing wasm file"
	@ GOOS=js GOARCH=wasm go build -o ./src/client/web/main/main.wasm ./src/client/web/main

clean:
	@ rm -dr dist || true
	@ rm -f ./src/client/web/main/main.wasm || true

create-dist:
	@ mkdir dist

run-server: clean create-dist compile-wasm
	$(info Running Server on Port :8081)
	@ go build -o ./dist/server ./src/server/main
	@ (cd ./; ./dist/server)

gotest:
	@ go test ./src/client ./src/ecs ./src/server

govet:
	@ go vet ./src/client ./src/ecs ./src/server ./src/game ./src/math
