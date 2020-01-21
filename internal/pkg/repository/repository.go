package repository

import (
	"database/sql"
	"fmt"
	"github.com/alexeynavarkin/technopark_db_proj/internal/pkg/consts"
	"github.com/alexeynavarkin/technopark_db_proj/internal/pkg/repository/cache"
	"github.com/alexeynavarkin/technopark_db_proj/internal/pkg/repository/seq"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db               *sqlx.DB
	users            cache.UserCache
	postsIDGenerator seq.Generator
}

func NewRepository(db *sqlx.DB) Repository {
	return Repository{
		db:               db,
		users:            cache.NewUserCache(),
		postsIDGenerator: seq.NewGenerator(),
	}
}

func (r *Repository) getOrder(desc bool) string {
	if desc {
		return " desc"
	}
	return ""
}
func (r *Repository) getLimit(limit int) string {
	if limit > 0 {
		return fmt.Sprintf(" limit %d", limit)
	}
	return ""
}

func (r *Repository) getSince(desc bool) string {
	if desc {
		return "<"
	}
	return ">"
}

func (r *Repository) handleError(err error) error {
	switch err {
	case sql.ErrNoRows:
		return consts.ErrNotFound
	default:
		return err
	}
}
