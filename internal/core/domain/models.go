package domain

type CreateCampaign struct {
	Name         string `json:"name" validate:"required"`
	Channel      string `json:"channel" validate:"required,oneof=sms whatsapp"`
	BaseTemplate string `json:"base_template" validate:"required"`
	ScheduledAt  string `json:"scheduled_at" validate:"omitempty,valid_timestamp"`
}

type CampaignsFilter struct {
	Status  string
	Channel string
}

type SendCampaign struct {
	CustomerIds []int64 `json:"customer_ids" validate:"required,min=1"`
}

type SendCampaignResult struct {
	CampaignID     int64  `json:"campaign_id"`
	MessagesQueued int32  `json:"messages_queued"`
	Status         string `json:"status"`
}

type PreviewMessage struct {
	CustomerID       int64  `json:"customer_id" validate:"required"`
	OverrideTemplate string `json:"override_template"`
}

type MinimalCustomer struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
}

type PreviewResponse struct {
	Message  string           `json:"rendered_message"`
	Template string           `json:"used_template"`
	Customer *MinimalCustomer `json:"customer"`
}
