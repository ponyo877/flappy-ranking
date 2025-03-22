.PHONY: dev
dev:
	npx wrangler dev

.PHONY: build
build:
	go run github.com/syumai/workers/cmd/workers-assets-gen@v0.28.1 -mode=go
	GOOS=js GOARCH=wasm go build -o ./build/app.wasm ./server/main.go

.PHONY: build-client
build-client:
	GOOS=js GOARCH=wasm go build -ldflags="-X 'main.serverURL='" -o ./static/main.wasm ./client/main.go ./client/http_client.go ./client/buttons.go
	gzip ./static/main.wasm

.PHONY: deploy
deploy:
	npx wrangler deploy

.PHONY: deploy-client
deploy-client:
	wrangler pages deploy ./static

.PHONY: create-db
create-db:
	npx wrangler d1 create flappy-ranking

.PHONY: init-db
init-db:
	npx wrangler d1 execute flappy-ranking --file=./storage/d1/schema.sql --remote

.PHONY: init-db-local
init-db-local:
	npx wrangler d1 execute flappy-ranking --file=./storage/d1/schema.sql --local

.PHONY: remove-db-local
remove-db-local:
	npx wrangler d1 execute flappy-ranking --file=./storage/d1/remove.sql --local
