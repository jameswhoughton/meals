
main_package_path = ./cmd/server
binary_name = meals_server

.PHONY: test
test:
	go test -v -race -buildvcs ./...

.PHONY: watch
watch:
	MEALS_DB_HOST="127.0.0.1" \
	MEALS_DB_PORT="8001" \
	MEALS_DB_USERNAME="root" \
	MEALS_DB_PASSWORD="" \
	MEALS_PORT="8000" \
	wgo run --file .gohtml --file .css --file .js ./$(main_package_path)/main.go

.PHONY: watch-tw
watch-tw:
	cd ./web && npx tailwindcss -i ./src/input.css -o ./static/main.css --watch

.PHONY: build-tw
build-tw:
	cd ./web && npx tailwindcss -i ./src/input.css -o ./static/main.css --minify

.PHONY: build
build:
	make build-tw && go build -o=$(binary_name) $(main_package_path)
	
.PHONY: build-image
build-image:
	docker build -t meals .

.PHONY: seed
seed:
	MEALS_DB_HOST="127.0.0.1" \
	MEALS_DB_PORT="8001" \
	MEALS_DB_USERNAME="root" \
	MEALS_DB_PASSWORD="" \
	MEALS_PORT="8000" \
	go run ./cmd/seeder/main.go -user-count=5
