package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// 引数の確認
	if len(os.Args) < 3 {
		fmt.Println("使用方法: go run cmd/migrate/main.go [環境名] [マイグレーション名]")
		fmt.Println("例: go run cmd/migrate/main.go dev add_user_table")
		fmt.Println("")
		fmt.Println("利用可能な環境:")
		fmt.Println("  dev    - ローカル開発環境 (postgres://postgres:postgres@localhost:5432/chat)")
		fmt.Println("  docker - Docker環境 (postgres://postgres:postgres@db:5432/chat)")
		os.Exit(1)
	}

	env := os.Args[1]
	migrationName := os.Args[2]

	// 環境の検証
	if env != "dev" && env != "docker" {
		fmt.Printf("エラー: 無効な環境名 '%s'\n", env)
		fmt.Println("利用可能な環境: dev, docker")
		os.Exit(1)
	}

	// マイグレーション名の検証
	if migrationName == "" {
		fmt.Println("エラー: マイグレーション名が指定されていません")
		os.Exit(1)
	}

	// 無効な文字をチェック
	if strings.ContainsAny(migrationName, " \t\n\r") {
		fmt.Println("エラー: マイグレーション名に空白文字は使用できません")
		os.Exit(1)
	}

	fmt.Printf("環境: %s\n", env)
	fmt.Printf("マイグレーション名: %s\n", migrationName)
	fmt.Println("")

	// 現在のディレクトリを取得
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get current directory: %v", err)
	}

	// バックエンドディレクトリに移動
	backendDir := filepath.Join(currentDir, "backend")
	if err := os.Chdir(backendDir); err != nil {
		log.Fatalf("failed to change directory to backend: %v", err)
	}

	// Atlasがインストールされているかチェック
	if !isAtlasInstalled() {
		fmt.Println("Atlasがインストールされていません。")
		fmt.Println("Dockerコンテナ内でAtlasをインストールしています...")

		// Dockerコンテナ内でAtlasをインストール
		if err := installAtlasInDocker(); err != nil {
			log.Fatalf("failed to install Atlas in Docker: %v", err)
		}
	}

	// マイグレーションファイルを生成
	fmt.Println("マイグレーションファイルを生成中...")
	if err := generateMigration(env, migrationName); err != nil {
		log.Fatalf("failed to generate migration: %v", err)
	}

	fmt.Println("")
	fmt.Println("✅ マイグレーションファイルが生成されました！")
	fmt.Println("")

	// 生成されたファイルを表示
	if err := showGeneratedFiles(); err != nil {
		fmt.Printf("⚠️  生成されたファイルの確認に失敗しました: %v\n", err)
	}

	fmt.Println("")
	fmt.Println("次のステップ:")
	fmt.Println("1. 生成されたマイグレーションファイルを確認")
	fmt.Println("2. 必要に応じてマイグレーションファイルを編集")
	fmt.Printf("3. マイグレーションを適用: docker-compose exec backend atlas migrate apply --env %s\n", env)
}

func isAtlasInstalled() bool {
	cmd := exec.Command("atlas", "version")
	return cmd.Run() == nil
}

func installAtlasInDocker() error {
	// Dockerコンテナ内でAtlasをインストール
	installCmd := exec.Command("docker-compose", "exec", "-T", "backend", "sh", "-c",
		"wget -O atlas.tar.gz https://github.com/ariga/atlas/releases/latest/download/atlas_linux_amd64.tar.gz && "+
			"tar -xzf atlas.tar.gz && "+
			"chmod +x atlas && "+
			"mv atlas /usr/local/bin/")

	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	return installCmd.Run()
}

func generateMigration(env, migrationName string) error {
	// Atlasコマンドを実行
	cmd := exec.Command("atlas", "migrate", "diff", migrationName,
		"--env", env,
		"--to", "file://schema",
		"--dir", "file://migrations")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func showGeneratedFiles() error {
	// migrationsディレクトリの内容を表示
	migrationsDir := "migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Println("生成されたファイル: なし")
		return nil
	}

	fmt.Println("生成されたファイル:")
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			fmt.Printf("  - %s\n", file.Name())
		}
	}

	return nil
}
