package app

import (
	"context"
	"fmt"
	"focus-dev-challenge/internal/adapters/repository"
	"focus-dev-challenge/internal/core/domain"
	"focus-dev-challenge/internal/core/ports"
	"reflect"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mwinyimoha/commons/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	repository ports.AppRepository
	validator  *validator.Validate
}

func NewService(r ports.AppRepository, v *validator.Validate) *Service {
	v.RegisterValidation("valid_timestamp", validTimestamp)

	return &Service{
		repository: r,
		validator:  v,
	}
}

func (svc *Service) AddCampaign(payload *domain.CreateCampaign) (*repository.Campaign, error) {
	if err := svc.validator.Struct(payload); err != nil {
		if verr, ok := err.(validator.ValidationErrors); ok {
			violations := errors.BuildViolations(verr)
			return nil, errors.NewValidationError(violations, "INVALID_REQUEST_DATA")
		}

		return nil, errors.WrapError(err, errors.InvalidArgument, "could not validate request data")
	}

	args := repository.CreateCampaignParams{
		Name:         payload.Name,
		Channel:      payload.Channel,
		Status:       "draft",
		BaseTemplate: payload.BaseTemplate,
	}
	if payload.ScheduledAt != "" {
		args.Status = "scheduled"

		ts, _ := time.Parse(time.RFC3339, payload.ScheduledAt)
		args.ScheduledAt = pgtype.Timestamp{
			Time:  ts,
			Valid: true,
		}
	}
	record, err := svc.repository.AddCampaign(&args)
	if err != nil {
		return nil, errors.WrapError(err, errors.Internal, "failed to create campaign")
	}

	return record, nil
}

func (svc *Service) ListCampaigns(pageNumber, pageSize int, filters *domain.CampaignsFilter) ([]*repository.ListCampaignsRow, error) {
	args := repository.ListCampaignsParams{
		PageNumber: pageNumber,
		PageSize:   int32(pageSize),
		Status:     filters.Status,
		Channel:    filters.Channel,
	}

	records, err := svc.repository.ListCampaigns(&args)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (svc *Service) RetrieveCampaign(campaignID int64) (*repository.GetCampaignRow, error) {
	return svc.repository.GetCampaign(campaignID)
}

func (svc *Service) PreviewMessage(campaignID int64, payload *domain.PreviewMessage) (*domain.PreviewResponse, error) {
	if err := svc.validator.Struct(payload); err != nil {
		if verr, ok := err.(validator.ValidationErrors); ok {
			violations := errors.BuildViolations(verr)
			return nil, errors.NewValidationError(violations, "INVALID_REQUEST_DATA")
		}

		return nil, errors.WrapError(err, errors.InvalidArgument, "could not validate request data")
	}

	var usedTemplate string
	if payload.OverrideTemplate != "" {
		usedTemplate = payload.OverrideTemplate
	} else {
		campaign, err := svc.repository.GetCampaign(campaignID)
		if err != nil {
			return nil, err
		}

		usedTemplate = campaign.BaseTemplate
	}

	customer, err := svc.repository.GetCustomer(payload.CustomerID)
	if err != nil {
		return nil, err
	}

	message := svc.renderTemplate(usedTemplate, customer)
	return &domain.PreviewResponse{
		Message:  message,
		Template: usedTemplate,
		Customer: &domain.MinimalCustomer{
			ID:        customer.ID,
			FirstName: customer.FirstName.String,
		},
	}, nil
}

func (svc *Service) SendCampaign(campaignID int64, payload *domain.SendCampaign) (*domain.SendCampaignResult, error) {
	if err := svc.validator.Struct(payload); err != nil {
		if verr, ok := err.(validator.ValidationErrors); ok {
			violations := errors.BuildViolations(verr)
			return nil, errors.NewValidationError(violations, "INVALID_REQUEST_DATA")
		}
		return nil, errors.WrapError(err, errors.InvalidArgument, "could not validate request data")
	}

	campaign, err := svc.repository.GetCampaign(campaignID)
	if err != nil {
		return nil, err
	}

	g, ctx := errgroup.WithContext(context.Background())
	sem := make(chan struct{}, 10)

	var queuedCount int32

	for _, customerID := range payload.CustomerIds {
		customerId := customerID

		g.Go(func() error {
			sem <- struct{}{}
			defer func() { <-sem }()

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			customer, err := svc.repository.GetCustomer(customerId)
			if err != nil {
				return err
			}

			message := svc.renderTemplate(campaign.BaseTemplate, customer)
			arg := repository.CreateOutboundMessageParams{
				CampaignID:      campaignID,
				CustomerID:      customerId,
				Status:          "pending",
				RenderedContent: message,
			}
			_, err = svc.repository.CreateOutboundMessage(&arg)
			if err != nil {
				return err
			}

			// TODO
			// Trigger Background Task

			atomic.AddInt32(&queuedCount, 1)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return &domain.SendCampaignResult{
			CampaignID:     campaignID,
			MessagesQueued: queuedCount,
			Status:         "sending",
		}, err
	}

	return &domain.SendCampaignResult{
		CampaignID:     campaignID,
		MessagesQueued: queuedCount,
		Status:         "sending",
	}, nil
}

func (svc *Service) renderTemplate(template string, data any) string {
	if template == "" || data == nil {
		return template
	}

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return template
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return template
	}

	re := regexp.MustCompile(`\{([^}]+)\}`)
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		fieldName := match[1 : len(match)-1]
		field := val.FieldByName(fieldName)
		if !field.IsValid() {
			return match
		}

		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				return ""
			}
			field = field.Elem()
		}

		if field.CanInterface() {
			if s, ok := field.Interface().(fmt.Stringer); ok {
				return s.String()
			}
		}

		if field.Kind() == reflect.Struct {
			// Handle common pgtype structs
			// pgtype.Text -> String
			if sf := field.FieldByName("String"); sf.IsValid() {
				return fmt.Sprint(sf.Interface())
			}

			// pgtype.Timestamp -> Time (format as RFC3339)
			if tf := field.FieldByName("Time"); tf.IsValid() {
				if tf.CanInterface() {
					if tt, ok := tf.Interface().(time.Time); ok {
						return tt.Format(time.RFC3339)
					}
					return fmt.Sprint(tf.Interface())
				}
			}
		}

		if field.CanInterface() {
			return fmt.Sprint(field.Interface())
		}

		return match
	})

	return result
}
