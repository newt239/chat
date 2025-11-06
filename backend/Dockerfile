# ビルドステージ
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 依存関係をインストール
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# バイナリをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o /server cmd/server/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o /migrate cmd/migrate/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o /reset cmd/reset/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o /seed cmd/seed/main.go

# 実行ステージ
FROM alpine:latest

# 必要なパッケージをインストール
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# ビルドステージからバイナリをコピー
COPY --from=builder /server .
COPY --from=builder /migrate .
COPY --from=builder /reset .
COPY --from=builder /seed .

# entスキーマをコピー（必要な場合）
COPY --from=builder /app/ent ./ent

# OpenAPI スキーマをコピー（本番用にイメージへ同梱）
RUN mkdir -p /root/backend/internal/openapi
COPY --from=builder /app/internal/openapi /root/backend/internal/openapi
ENV OPENAPI_SPEC_PATH=backend/internal/openapi/openapi.yaml

# ポートを公開
EXPOSE 8080

# サーバーを起動
CMD ["./server"]
