package ports

import (
	"focus-dev-challenge/internal/adapters/repository"
	"focus-dev-challenge/internal/core/domain"
)

type AppService interface {
	AddCampaign(payload *domain.CreateCampaign) (*repository.Campaign, error)
	ListCampaigns(pageNumber, pageSize int, filters *domain.CampaignsFilter) ([]*repository.ListCampaignsRow, error)
	RetrieveCampaign(campaignID int64) (*repository.GetCampaignRow, error)
	PreviewMessage(campaignID int64, payload *domain.PreviewMessage) (*domain.PreviewResponse, error)
	SendCampaign(campaignID int64, payload *domain.SendCampaign) (*domain.SendCampaignResult, error)
}
