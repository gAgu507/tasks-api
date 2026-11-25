package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gAgu507/tasks-api/internal/config"
	"github.com/gAgu507/tasks-api/internal/handlers"
	"github.com/gAgu507/tasks-api/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("godotenv:", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("connect to db: ", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("ping db: ", err)
	}

	pg := store.NewPostgres(pool)
	srv := handlers.NewServer(pg, pg, []byte(cfg.JWTSecret))

	log.Println("Server is running on http://localhost:" + cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, srv.Routes()))
}
