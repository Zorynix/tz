package handler

import (
	"net/http"

	"subscription-service/internal/domain"
	"subscription-service/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SubscriptionHandler struct {
	service service.SubscriptionService
	logger  *zap.Logger
}

func NewSubscriptionHandler(service service.SubscriptionService, logger *zap.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		logger:  logger,
	}
}

// CreateSubscription godoc
// @Summary Create a new subscription
// @Description Create a new subscription record
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body domain.CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} domain.Subscription
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	h.logger.Info("handler: create subscription request")

	var req domain.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create subscription", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("subscription created successfully", zap.String("id", subscription.ID.String()))
	c.JSON(http.StatusCreated, subscription)
}

// GetSubscription godoc
// @Summary Get subscription by ID
// @Description Get subscription details by ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Success 200 {object} domain.Subscription
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	h.logger.Info("handler: get subscription request")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("invalid subscription id", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
		return
	}

	subscription, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to get subscription", zap.String("id", id.String()), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	h.logger.Info("subscription retrieved successfully", zap.String("id", id.String()))
	c.JSON(http.StatusOK, subscription)
}

// UpdateSubscription godoc
// @Summary Update subscription
// @Description Update subscription by ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Param subscription body domain.UpdateSubscriptionRequest true "Subscription update data"
// @Success 200 {object} domain.Subscription
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	h.logger.Info("handler: update subscription request")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("invalid subscription id", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
		return
	}

	var req domain.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update subscription", zap.String("id", id.String()), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("subscription updated successfully", zap.String("id", id.String()))
	c.JSON(http.StatusOK, subscription)
}

// DeleteSubscription godoc
// @Summary Delete subscription
// @Description Delete subscription by ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	h.logger.Info("handler: delete subscription request")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Error("invalid subscription id", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.logger.Error("failed to delete subscription", zap.String("id", id.String()), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("subscription deleted successfully", zap.String("id", id.String()))
	c.Status(http.StatusNoContent)
}

// ListSubscriptions godoc
// @Summary List subscriptions
// @Description List subscriptions with optional filters
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param user_id query string false "User ID filter"
// @Param service_name query string false "Service name filter"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	h.logger.Info("handler: list subscriptions request")

	var req domain.ListSubscriptionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("failed to bind query", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscriptions, total, err := h.service.List(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to list subscriptions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("subscriptions listed successfully", zap.Int("count", len(subscriptions)), zap.Int64("total", total))
	c.JSON(http.StatusOK, gin.H{
		"data":   subscriptions,
		"total":  total,
		"limit":  req.Limit,
		"offset": req.Offset,
	})
}

// CalculateTotalCost godoc
// @Summary Calculate total cost
// @Description Calculate total cost of subscriptions for a period
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param user_id query string false "User ID filter"
// @Param service_name query string false "Service name filter"
// @Param start_date query string true "Start date (MM-YYYY)"
// @Param end_date query string true "End date (MM-YYYY)"
// @Success 200 {object} domain.TotalCostResponse
// @Failure 400 {object} map[string]interface{}
// @Router /subscriptions/total-cost [get]
func (h *SubscriptionHandler) CalculateTotalCost(c *gin.Context) {
	h.logger.Info("handler: calculate total cost request")

	var req domain.TotalCostRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("failed to bind query", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.CalculateTotalCost(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to calculate total cost", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("total cost calculated successfully", zap.Int("total_cost", result.TotalCost))
	c.JSON(http.StatusOK, result)
}
