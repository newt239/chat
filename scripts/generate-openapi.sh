#!/bin/bash

# OpenAPIスキーマからGoの型とサーバースタブを生成するスクリプト
# 使用方法: ./scripts/generate-openapi.sh

set -e

# プロジェクトルートに移動
cd "$(dirname "$0")/.."

# bundled.yamlが存在しない場合は生成
if [ ! -f "openapi/bundled.yaml" ]; then
    echo "bundled.yamlが見つかりません。生成します..."
    pnpm run openapi:bundle
fi

# oapi-codegenがインストールされているか確認
if ! command -v oapi-codegen &> /dev/null; then
    echo "oapi-codegenがインストールされていません。インストールします..."
    go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
    
    # GOPATH/binをPATHに追加
    if [ -n "$GOPATH" ]; then
        export PATH="$GOPATH/bin:$PATH"
    elif [ -d "$HOME/go/bin" ]; then
        export PATH="$HOME/go/bin:$PATH"
    fi
fi

# 出力ディレクトリを作成
mkdir -p backend/internal/openapi_gen

# コード生成
echo "OpenAPIコードを生成中..."
oapi-codegen \
  -generate types,server \
  -package openapi \
  -o backend/internal/openapi_gen/openapi.gen.go \
  openapi/bundled.yaml

echo "✅ コード生成が完了しました: backend/internal/openapi_gen/openapi.gen.go"

