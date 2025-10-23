#!/bin/bash

# カラー定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}🧹 プロジェクトのクリーンアップを開始します...${NC}"

# プロジェクトルートに移動
cd "$(dirname "$0")/.."

# 削除対象のファイル・ディレクトリを定義
TARGETS=(
    ".turbo"
    "frontend/.turbo"
    "backend/.turbo"
    "node_modules"
    "frontend/node_modules"
    "pnpm-lock.yaml"
    "frontend/dist"
    "build"
    "coverage"
    "frontend/.vite"
    "frontend/.tanstack"
    ".cache"
    "*.log"
    ".DS_Store"
)

# 各ターゲットを削除
for target in "${TARGETS[@]}"; do
    if [[ "$target" == "*.log" || "$target" == ".DS_Store" ]]; then
        # ワイルドカードパターンの場合
        find . -name "$target" -type f -delete 2>/dev/null
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✅ $target ファイルを削除しました${NC}"
        fi
    else
        # 通常のファイル・ディレクトリの場合
        if [ -e "$target" ] || [ -d "$target" ]; then
            rm -rf "$target"
            echo -e "${GREEN}✅ $target を削除しました${NC}"
        else
            echo -e "${YELLOW}⚠️  $target が見つかりませんでした${NC}"
        fi
    fi
done

# pnpmキャッシュもクリア
echo -e "${YELLOW}🗑️  pnpmキャッシュをクリア中...${NC}"
pnpm store prune 2>/dev/null
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ pnpmキャッシュをクリアしました${NC}"
else
    echo -e "${YELLOW}⚠️  pnpmキャッシュのクリアに失敗しました（pnpmがインストールされていない可能性があります）${NC}"
fi

echo -e "${GREEN}🎉 クリーンアップが完了しました！${NC}"
echo -e "${YELLOW}💡 依存関係を再インストールするには: pnpm install${NC}"
