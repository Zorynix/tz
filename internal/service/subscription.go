package service

import (
	"context"
	"errors"

	"subscription-service/internal/domain"
	"subscription-service/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SubscriptionService interface {
	Create(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	Update(ctx context.Context, id uuid.UUID, req *domain.UpdateSubscriptionRequest) (*domain.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *domain.ListSubscriptionsRequest) ([]*domain.Subscription, int64, error)
	CalculateTotalCost(ctx context.Context, req *domain.TotalCostRequest) (*domain.TotalCostResponse, error)
}

type subscriptionService struct {
	repo   repository.SubscriptionRepository
	logger *zap.Logger
}

func NewSubscriptionService(repo repository.SubscriptionRepository, logger *zap.Logger) SubscriptionService {
	return &subscriptionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *subscriptionService) Create(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
	s.logger.Info("service: creating subscription", zap.String("service_name", req.ServiceName))

	if err := s.validateDateFormat(req.StartDate); err != nil {
		s.logger.Error("invalid start date format", zap.String("start_date", req.StartDate), zap.Error(err))
		return nil, err
	}

	if req.EndDate != nil {
		if err := s.validateDateFormat(*req.EndDate); err != nil {
			s.logger.Error("invalid end date format", zap.String("end_date", *req.EndDate), zap.Error(err))
			return nil, err
		}

		if *req.EndDate < req.StartDate {
			s.logger.Error("end date must be after start date")
			return nil, errors.New("end date must be after start date")
		}
	}

	return s.repo.Create(ctx, req)
}

func (s *subscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	s.logger.Info("service: getting subscription", zap.String("id", id.String()))
	return s.repo.GetByID(ctx, id)
}

func (s *subscriptionService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateSubscriptionRequest) (*domain.Subscription, error) {
	s.logger.Info("service: updating subscription", zap.String("id", id.String()))

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("subscription not found", zap.String("id", id.String()), zap.Error(err))
		return nil, errors.New("subscription not found")
	}

	if req.StartDate != nil {
		if err := s.validateDateFormat(*req.StartDate); err != nil {
			s.logger.Error("invalid start date format", zap.String("start_date", *req.StartDate), zap.Error(err))
			return nil, err
		}
	}

	if req.EndDate != nil {
		if err := s.validateDateFormat(*req.EndDate); err != nil {
			s.logger.Error("invalid end date format", zap.String("end_date", *req.EndDate), zap.Error(err))
			return nil, err
		}
	}

	return s.repo.Update(ctx, id, req)
}

func (s *subscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	s.logger.Info("service: deleting subscription", zap.String("id", id.String()))

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("subscription not found", zap.String("id", id.String()), zap.Error(err))
		return errors.New("subscription not found")
	}

	return s.repo.Delete(ctx, id)
}

func (s *subscriptionService) List(ctx context.Context, req *domain.ListSubscriptionsRequest) ([]*domain.Subscription, int64, error) {
	s.logger.Info("service: listing subscriptions")

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	filter := &repository.ListSubscriptionsFilter{
		ServiceName: req.ServiceName,
		Limit:       req.Limit,
		Offset:      req.Offset,
	}

	if req.UserID != nil && *req.UserID != "" {
		userID, err := uuid.Parse(*req.UserID)
		if err != nil {
			s.logger.Error("invalid user_id format", zap.String("user_id", *req.UserID), zap.Error(err))
			return nil, 0, errors.New("invalid user_id format")
		}
		filter.UserID = &userID
	}

	return s.repo.List(ctx, filter)
}

func (s *subscriptionService) CalculateTotalCost(ctx context.Context, req *domain.TotalCostRequest) (*domain.TotalCostResponse, error) {
	s.logger.Info("service: calculating total cost")

	if err := s.validateDateFormat(req.StartDate); err != nil {
		s.logger.Error("invalid start date format", zap.String("start_date", req.StartDate), zap.Error(err))
		return nil, err
	}

	if err := s.validateDateFormat(req.EndDate); err != nil {
		s.logger.Error("invalid end date format", zap.String("end_date", req.EndDate), zap.Error(err))
		return nil, err
	}

	filter := &repository.TotalCostFilter{
		ServiceName: req.ServiceName,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	if req.UserID != nil && *req.UserID != "" {
		userID, err := uuid.Parse(*req.UserID)
		if err != nil {
			s.logger.Error("invalid user_id format", zap.String("user_id", *req.UserID), zap.Error(err))
			return nil, errors.New("invalid user_id format")
		}
		filter.UserID = &userID
	}

	totalCost, err := s.repo.CalculateTotalCost(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &domain.TotalCostResponse{TotalCost: totalCost}, nil
}

func (s *subscriptionService) validateDateFormat(date string) error {
	if len(date) != 10 {
		return errors.New("date must be in YYYY-MM-DD format")
	}

	if date[4] != '-' || date[7] != '-' {
		return errors.New("date must be in YYYY-MM-DD format")
	}

	year := date[:4]
	month := date[5:7]
	day := date[8:]

	if len(year) != 4 || len(month) != 2 || len(day) != 2 {
		return errors.New("date must be in YYYY-MM-DD format")
	}

	return nil
}
