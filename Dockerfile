# standard dockerfile for the go module
FROM golang:1.22.3-alpine AS build

WORKDIR /app

COPY go.mod .

COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o /go/bin/bfe

FROM debian:bookworm-slim

# Make port 8888 available to the world outside this container
EXPOSE 8082

COPY --from=build /go/bin/bfe /go/bin/bfe

CMD ["/go/bin/bfe"]

# Path: main.go

