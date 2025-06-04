
main_package_path = ./cmd/server
binary_name = meals_server

.PHONY: test
test:
	go test -v -race -buildvcs ./...

.PHONY: watch
watch:
	MEALS_DSN="root@tcp(127.0.0.1:8001)/meals?parseTime=true" \
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
