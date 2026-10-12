package main

import (
	"cellar-app/database"
	"cellar-app/handler"
	"cellar-app/llm"
	"cellar-app/service"
	"log"
	"os"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("DB接続に失敗しました: %v", err)
	}
	log.Println("Database connected")

	if err := database.Migrate(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migration completed")

	provider, err := llm.NewWineInfoProvider()
	if err != nil {
		log.Fatalf("LLM provider initialization failed: %v", err)
	}

	svc := service.NewService(db, provider)

	if err := svc.StartDailySnapshotScheduler(); err != nil {
		log.Printf("Warning: Failed to start snapshot scheduler: %v\n", err)
	}

	h := &handler.Handler{Service: svc}
	r := SetupRouter(h)

	certFile := os.Getenv("SSL_CERT_FILE")
	keyFile := os.Getenv("SSL_KEY_FILE")

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
