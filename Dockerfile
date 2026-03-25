FROM golang:1.26 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /bin/server /app/server
EXPOSE 8080
CMD ["/app/server"]
