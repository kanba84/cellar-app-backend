package main

import (
	"cellar-app/handler"
	"cellar-app/service"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func main() {
	// 環境変数からDB接続情報を取得（デフォルト値あり）
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "cellar_app"
	}
	password := os.Getenv("DB_PASSWORD") // パスワードが設定されていない場合は空文字列
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "cellar_app"
	}

	// データベース接続文字列を構築
	dbURL := "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbname

	// データベース接続
	var err error
	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("DB接続に失敗しました: %v", err)
	}
	defer dbPool.Close()

	// ルーター設定してサーバー起動
	svc := &service.Service{Pool: dbPool}
	h := &handler.Handler{Service: svc}
	r := SetupRouter(h)

	// SSL証明書・鍵のパス
	certFile := os.Getenv("SSL_CERT_FILE")
	keyFile := os.Getenv("SSL_KEY_FILE")
	fmt.Println("certFile:", certFile, "keyFile:", keyFile)
	if certFile != "" && keyFile != "" {
		// HTTPSでサーバー起動
		fmt.Println("HTTPS server starting on port 8443")
		log.Println("HTTPS server starting on port 8443")
		if err := r.RunTLS(":8443", certFile, keyFile); err != nil {
			log.Fatalf("failed to starting the server: %v", err)
		}
	} else {
		// HTTPでサーバー起動
		log.Println("HTTP server starting on port 8080")
		fmt.Println("HTTP server starting on port 8080")
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("failed to starting the server: %v", err)
		}
	}
}
