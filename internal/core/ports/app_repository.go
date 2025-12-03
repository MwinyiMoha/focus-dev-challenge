package ports

import (
	"context"
	"focus-dev-challenge/internal/adapters/repository"
)

type AppRepository interface {
	Close() error
	ExecTx(ctx context.Context, fn func(*repository.Queries) error) error

	AddCampaign(arg *repository.CreateCampaignParams) (*repository.Campaign, error)
	ListCampaigns(arg *repository.ListCampaignsParams) ([]*repository.ListCampaignsRow, error)
	GetCampaign(ID int64) (*repository.GetCampaignRow, error)

	GetCustomer(ID int64) (*repository.Customer, error)

	CreateOutboundMessage(arg *repository.CreateOutboundMessageParams) (*repository.OutboundMessage, error)
}
