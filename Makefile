version := 0.0.1-SNAPSHOT

all: clean build run

docker: clean build dockerBuild dockerRun

build:
	mkdir ./build
	CGO_ENABLED=0 go build -o ./build/spellscan-card-loader

run:
	chmod +x ./build/spellscan-card-loader
	./build/spellscan-card-loader

dockerBuild:
	docker build \
		-t spellscan-card-loader:$version \
		-t spellscan-card-loader:latest \
		.

dockerRun:
	docker run --rm \
		-e DB_DSN=$(DB_DSN) \
		-e MEILI_URL=$(MEILI_URL) \
		-e MEILI_API_KEY=$(MEILI_API_KEY) \
		spellscan-card-loader:latest

clean:
	rm -r build || true
	rm -r tmp || true