// @title           Recommendation System API
// @version         1.0
// @description     Система рекомендаций загрузки производственного оборудования
// @BasePath        /

package main

import (
	"context"
	"log"
	"net/http"

	"recsys-backend/internal/config"
	"recsys-backend/internal/httpapi"
	"recsys-backend/internal/service"
	"recsys-backend/internal/storage"

	"github.com/joho/godotenv"

	_ "recsys-backend/docs" // swag init создаст пакет docs
)

func main() {
	_ = godotenv.Load()

	cfg := config.FromEnv()

	ctx := context.Background()
	db, err := storage.NewDB(ctx, cfg.DB)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repos := storage.NewRepos(db)
	planner := service.NewPlanner(repos) // пока эвристика-минимум
	h := httpapi.NewHandlers(repos, planner)

	r := httpapi.NewRouter(h)

	log.Println("Listening on", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		log.Fatal(err)
	}
}
