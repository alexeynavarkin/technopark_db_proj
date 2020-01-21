package app

import (
	"fmt"
	"github.com/alexeynavarkin/technopark_db_proj/internal/pkg/handler"
	"github.com/alexeynavarkin/technopark_db_proj/internal/pkg/repository"
	"github.com/alexeynavarkin/technopark_db_proj/internal/pkg/usecase"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/valyala/fasthttp"
	"log"
	"os"
)

const PORT = "5000"

func Start() {
	db, err := newDB()
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewRepository(db)
	u := usecase.NewUsecase(repo)
	h := handler.NewHandler(u)

	fmt.Println("listening port " + PORT)
	log.Fatal(fasthttp.ListenAndServe(":"+PORT, h.GetHandleFunc()))
}

func newDB() (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", os.Getenv("POSTGRES_DSN"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(8)
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
