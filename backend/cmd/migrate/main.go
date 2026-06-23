package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/wangxintong/yijing/backend/internal/config"
	"github.com/wangxintong/yijing/backend/internal/db"
	"github.com/wangxintong/yijing/backend/internal/migrate"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	sqlDir, err := migrate.ResolveSQLDir(cfg.SQLDir)
	if err != nil {
		log.Fatalf("resolve sql dir: %v", err)
	}
	log.Printf("using sql dir: %s", sqlDir)

	mysqlDB, err := db.Connect(cfg.DSN())
	if err != nil {
		log.Fatalf("connect mysql: %v", err)
	}
	defer mysqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	runner := migrate.NewRunner(mysqlDB, sqlDir)
	if err := runner.Run(ctx); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("migration completed")
	os.Exit(0)
}
