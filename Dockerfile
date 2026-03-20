FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o service .

FROM scratch
COPY --from=builder /app/service /service
ENTRYPOINT ["/service"]
