package api

import (
	"encoding/json"
	"focus-dev-challenge/internal/core/domain"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mwinyimoha/commons/pkg/errors"
)

func (r *Router) GetCampaigns(c *gin.Context) {
	pageNumber := 1
	pageSize := 10

	if v, err := strconv.Atoi(c.Query("page_number")); err == nil {
		pageNumber = v
	}

	if v, err := strconv.Atoi(c.Query("page_size")); err == nil {
		if v > 0 && v <= 100 {
			pageSize = v
		}
	}

	filter := domain.CampaignsFilter{
		Status:  c.Query("status"),
		Channel: c.Query("channel"),
	}

	records, err := r.service.ListCampaigns(pageNumber, pageSize, &filter)
	if err != nil {
		if cerr, ok := err.(*errors.Error); ok {
			code, detail := cerr.HTTPStatus()
			c.JSON(code, detail)
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	totalCount := int64(0)
	if len(records) > 0 {
		totalCount = records[0].TotalCount
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	res := gin.H{
		"data": records,
		"pagination": gin.H{
			"page":        pageNumber,
			"page_size":   pageSize,
			"total_count": totalCount,
			"total_pages": totalPages,
		},
	}

	c.JSON(http.StatusOK, res)
}

func (r *Router) CreateCampaign(c *gin.Context) {
	var data domain.CreateCampaign
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	campaign, err := r.service.AddCampaign(&data)
	if err != nil {
		if cerr, ok := err.(*errors.Error); ok {
			code, detail := cerr.HTTPStatus()
			c.JSON(code, detail)
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, campaign)
}

func (r *Router) GetCampaign(c *gin.Context) {
	ID := c.Param("id")
	campaignID, err := strconv.Atoi(ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	campaign, err := r.service.RetrieveCampaign(int64(campaignID))
	if err != nil {
		if cerr, ok := err.(*errors.Error); ok {
			code, detail := cerr.HTTPStatus()
			c.JSON(code, detail)
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	var stats any
	if err := json.Unmarshal(campaign.Stats, &stats); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	result := gin.H{
		"id":            campaign.ID,
		"name":          campaign.Name,
		"channel":       campaign.Channel,
		"status":        campaign.Status,
		"base_template": campaign.BaseTemplate,
		"scheduled_at":  campaign.ScheduledAt,
		"created_at":    campaign.CreatedAt,
		"stats":         stats,
	}
	c.JSON(http.StatusOK, result)
}

func (r *Router) SendCampaign(c *gin.Context) {
	ID := c.Param("id")
	campaignID, err := strconv.Atoi(ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var data domain.SendCampaign
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	result, err := r.service.SendCampaign(int64(campaignID), &data)
	if err != nil {
		if cerr, ok := err.(*errors.Error); ok {
			code, detail := cerr.HTTPStatus()
			c.JSON(code, detail)
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (r *Router) Preview(c *gin.Context) {
	ID := c.Param("id")
	campaignID, err := strconv.Atoi(ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var data domain.PreviewMessage
	if err := c.ShouldBind(&data); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	result, err := r.service.PreviewMessage(int64(campaignID), &data)
	if err != nil {
		if cerr, ok := err.(*errors.Error); ok {
			code, detail := cerr.HTTPStatus()
			c.JSON(code, detail)
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
