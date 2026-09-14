FROM golang:1.27 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /bin/once ./cmd

FROM debian:trixie-slim AS debian

RUN apt-get update && apt-get install -y --no-install-recommends ffmpeg curl && rm -rf /var/lib/apt/lists/*

WORKDIR /api

COPY --from=build /bin/once /bin/once
CMD [ "/bin/once" ]
