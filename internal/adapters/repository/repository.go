package repository

import (
	"context"
	"focus-dev-challenge/internal/config"
	"time"

	"github.com/jackc/pgx/v5"
)

type Repository struct {
	*Queries
	db        *pgx.Conn
	dbTimeout time.Duration // seconds
}

func NewRepository(cfg *config.Config) (*Repository, error) {
	defaultTimeout := time.Duration(cfg.DefaultTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:        conn,
		Queries:   New(conn),
		dbTimeout: defaultTimeout,
	}, nil
}

func (r *Repository) Close() error {
	ctx, cancel := r.getContext()
	defer cancel()

	return r.db.Close(ctx)
}

func (r *Repository) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), r.dbTimeout)
}
