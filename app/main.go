package main

import (
	"cellar-app/handler"
	"cellar-app/service"
	"context"
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
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("サーバーの起動に失敗しました: %v", err)
	}
}
