FROM golang:1.26-alpine AS base

WORKDIR /app

RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download

FROM base AS dev

# install.sh はチェックサム検証に失敗することがあるため公式イメージから取得する
COPY --from=golangci/golangci-lint:v2.12.2-alpine /usr/bin/golangci-lint /usr/local/bin/golangci-lint
# @latest は Go の要求バージョンが上がるとビルドが壊れる
RUN go install github.com/air-verse/air@v1.66.0

EXPOSE 8080

CMD ["air", "-c", "air.toml"]

FROM base AS builder

COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server && \
    CGO_ENABLED=0 go build -o /reindex ./cmd/reindex

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /server /reindex ./

EXPOSE 8080

CMD ["./server"]
