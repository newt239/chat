FROM golang:1.27-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server cmd/server/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o /reindex cmd/reindex/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /server .
COPY --from=builder /reindex .

EXPOSE 8080

CMD ["./server"]
