
main_package_path = ./cmd/server
binary_name = meals_server

.PHONY: test
test:
	go test -v -race -buildvcs ./...

.PHONY: watch
watch:
	wgo run --file .gohtml --file .css --file .js ./$(main_package_path)/main.go

.PHONY: watch-tw
watch-tw:
	cd ./frontend && npx tailwindcss -i ../web/src/input.css -o ../web/static/main.css --watch

.PHONY: build-tw
build-tw:
	cd ./frontend && npx tailwindcss -i ../web/src/input.css -o ../web/static/main.css --minify

.PHONY: build
build:
	make build-tw && go build -o=$(binary_name) $(main_package_path)
	
.PHONY: build-image
build-image:
	docker build -t meals .
