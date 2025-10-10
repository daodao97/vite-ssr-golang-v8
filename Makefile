.PHONY: clean build

APP_NAME = server
BUILD_DIR = $(PWD)/build

up:
	docker compose up -d

down:
	docker compose down

clean:
	rm -rf ./build
	rm -rf ./dist
	rm -rf ./webssr/dist

build:
	# build the backend
	go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME) main.go

web:
	cd webssr && pnpm build

webdev:
	pnpm --dir webssr run dev	

typecheck:
	pnpm --dir webssr run typecheck