version := 0.0.6

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
		-t ghcr.io/murilo-bracero/spellscan-card-loader:$(version) \
		-t ghcr.io/murilo-bracero/spellscan-card-loader:latest \
		.

dockerPublish:
	docker push ghcr.io/murilo-bracero/spellscan-card-loader:$(version)
	docker push ghcr.io/murilo-bracero/spellscan-card-loader:latest

dockerRun:
	docker run --rm \
		-e DB_DSN=$(DB_DSN) \
		-e MEILI_URL=$(MEILI_URL) \
		-e MEILI_API_KEY=$(MEILI_API_KEY) \
		-e USE_RELEASE_DATE_REFERENCE=$(USE_RELEASE_DATE_REFERENCE) \
		ghcr.io/murilo-bracero/spellscan-card-loader:latest

clean:
	rm -r build || true
	rm -r tmp || true

showVersion:
	@echo $(version)