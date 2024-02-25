all: clean build run

build:
	mkdir ./build
	go build -o ./build/spellscan-card-loader

run:
	source .env || true
	chmod +x ./build/spellscan-card-loader
	./build/spellscan-card-loader

clean:
	rm -r build