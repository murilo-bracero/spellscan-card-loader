FROM golang:1.21-alpine3.20 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN apk add --no-cache make

COPY . .

RUN make build

FROM alpine:3.20 as runner

ENV GROUP=spellscancardloader
ENV USER=spellscancardloaderapp
ENV UID=10001

RUN addgroup -S ${GROUP} \
    && adduser -S -u ${UID} -g ${GROUP} ${USER}

WORKDIR /app

COPY --chown=${USER}:${GROUP} --from=builder /app/build/spellscan-card-loader spellscan-card-loader

USER ${USER}

CMD ["./spellscan-card-loader"]