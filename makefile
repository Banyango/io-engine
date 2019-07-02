
compile-wasm:
	echo "Builing wasm file"
	@ GOOS=js GOARCH=wasm go build -o ./dist/app/main/main.wasm ./src/client/web/main
	$(info wasm comiled -> /src/client/web/main/main.wasm)

copy-web-template:
	@ cp -a ./src/client/web/main/template/. ./dist/app/main/

clean:
	@ rm -dr dist || true
	@ rm -f ./src/client/web/main/main.wasm || true

create-dist:
	@ mkdir dist

run-server: clean create-dist compile-wasm copy-web-template
	$(info Running Server on Port :8081)
	@ go build -o ./dist/server ./src/server/main
	@ cp ./game.json ./dist/
	@ (cd ./dist; ./server)

build: clean create-dist compile-wasm copy-web-template
	$(info Building...)
	@ CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o ./dist/server ./src/server/main
	@ chmod 777 ./dist/server
	@ cp ./game.json ./dist/

gotest:
	@ go test ./src/client ./src/ecs ./src/server

govet:
	@ go vet ./src/client ./src/ecs ./src/server ./src/game ./src/math
