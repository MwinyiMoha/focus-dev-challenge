package repository

import (
	"context"
	"fmt"
	"focus-dev-challenge/internal/config"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mwinyimoha/commons/pkg/errors"
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

func (r *Repository) ExecTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	q := New(tx)

	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("transaction err: %v, rollback err: %v", err, rbErr)
		}

		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) AddCampaign(arg *CreateCampaignParams) (*Campaign, error) {
	ctx, cancel := r.getContext()
	defer cancel()

	record, err := r.Queries.CreateCampaign(ctx, arg)
	if err != nil {
		return nil, errors.WrapError(err, errors.Internal, "SAVE_CAMPAIGN_ERROR")
	}

	return record, nil
}

func (r *Repository) ListCampaigns(arg *ListCampaignsParams) ([]*ListCampaignsRow, error) {
	ctx, cancel := r.getContext()
	defer cancel()

	records, err := r.Queries.ListCampaigns(ctx, arg)
	if err != nil {
		return nil, errors.WrapError(err, errors.Internal, "FETCH_CAMPAIGNS_ERROR")
	}

	return records, nil
}

func (r *Repository) GetCampaign(ID int64) (*GetCampaignRow, error) {
	ctx, cancel := r.getContext()
	defer cancel()

	record, err := r.Queries.GetCampaign(ctx, ID)
	if err != nil {
		return nil, errors.WrapError(err, errors.Internal, "FETCH_CAMPAIGN_ERROR")
	}

	return record, nil
}

func (r *Repository) GetCustomer(ID int64) (*Customer, error) {
	ctx, cancel := r.getContext()
	defer cancel()

	record, err := r.Queries.GetCustomerById(ctx, ID)
	if err != nil {
		return nil, errors.WrapError(err, errors.Internal, "FETCH_CUSTOMER_ERROR")
	}

	return record, nil
}

func (r *Repository) CreateOutboundMessage(arg *CreateOutboundMessageParams) (*OutboundMessage, error) {
	ctx, cancel := r.getContext()
	defer cancel()

	record, err := r.Queries.CreateOutboundMessage(ctx, arg)
	if err != nil {
		return nil, errors.WrapError(err, errors.Internal, "CREATE_OUTBOUND_MESSAGE_ERROR")
	}

	return record, nil
}

func (r *Repository) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), r.dbTimeout)
}
