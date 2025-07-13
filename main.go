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
	// 環境変数からホスト名を取得
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost" // デフォルト値
	}

	// データベース接続文字列を構築
	dbURL := "postgres://cellar_app:P1n0tN01r@" + host + ":5432/cellar_app"

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
