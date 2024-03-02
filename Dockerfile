FROM alpine:3.16

WORKDIR /app

COPY build/spellscan-card-loader spellscan-card-loader

CMD ["./spellscan-card-loader"]