package ports

import "focus-dev-challenge/internal/adapters/repository"

type AppRepository interface {
	Close() error

	AddCampaign(arg *repository.CreateCampaignParams) (*repository.Campaign, error)
	ListCampaigns(arg *repository.ListCampaignsParams) ([]*repository.ListCampaignsRow, error)
	GetCampaign(ID int64) (*repository.GetCampaignRow, error)

	GetCustomer(ID int64) (*repository.Customer, error)

	CreateOutboundMessage(arg *repository.CreateOutboundMessageParams) (*repository.OutboundMessage, error)
}
