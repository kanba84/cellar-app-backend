package main

import (
	"cellar-app/handler"
	"cellar-app/service"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 環境変数からDB接続情報を取得
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "cellar_app"
	}

	password := os.Getenv("DB_PASSWORD")

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "cellar_app"
	}

	// GORM用DSN
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port,
	)

	// DB接続
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB接続に失敗しました: %v", err)
	}

	fmt.Println("DB接続成功")

	// ServiceにDBを渡す（ここが最重要）
	svc := service.NewService(db)

	// 毎日0:00にスナップショットを作成するスケジューラーを開始
	if err := svc.StartDailySnapshotScheduler(); err != nil {
		log.Printf("Warning: Failed to start snapshot scheduler: %v\n", err)
	}

	h := &handler.Handler{Service: svc}
	r := SetupRouter(h)

	// SSL証明書・鍵
	certFile := os.Getenv("SSL_CERT_FILE")
	keyFile := os.Getenv("SSL_KEY_FILE")

	fmt.Println("certFile:", certFile, "keyFile:", keyFile)

	if certFile != "" && keyFile != "" {
		log.Println("HTTPS server starting on port 8443")
		if err := r.RunTLS(":8443", certFile, keyFile); err != nil {
			log.Fatalf("failed to starting the server: %v", err)
		}
	} else {
		log.Println("HTTP server starting on port 8080")
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("failed to starting the server: %v", err)
		}
	}
}
