FROM golang:1.25.5-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/stock-api ./cmd/api

FROM alpine:3.20
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=build /app/stock-api /app/stock-api
USER appuser

EXPOSE 8080
CMD ["/app/stock-api"]
