package repository

import (
	"context"

	"subscription-service/internal/domain"
	"subscription-service/internal/repository/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type ListSubscriptionsFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	Limit       int
	Offset      int
}

type TotalCostFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	StartDate   string
	EndDate     string
}

type SubscriptionRepository interface {
	Create(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	Update(ctx context.Context, id uuid.UUID, req *domain.UpdateSubscriptionRequest) (*domain.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *ListSubscriptionsFilter) ([]*domain.Subscription, int64, error)
	CalculateTotalCost(ctx context.Context, filter *TotalCostFilter) (int, error)
}

type subscriptionRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
	logger  *zap.Logger
}

func NewSubscriptionRepository(db *pgxpool.Pool, logger *zap.Logger) SubscriptionRepository {
	return &subscriptionRepository{
		db:      db,
		queries: sqlc.New(db),
		logger:  logger,
	}
}

func (r *subscriptionRepository) Create(ctx context.Context, req *domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
	r.logger.Info("creating subscription", zap.String("service_name", req.ServiceName), zap.String("user_id", req.UserID.String()))

	userIDPgtype := pgtype.UUID{}
	if err := userIDPgtype.Scan(req.UserID.String()); err != nil {
		r.logger.Error("failed to convert user_id", zap.Error(err))
		return nil, err
	}

	var endDate pgtype.Text
	if req.EndDate != nil {
		endDate = pgtype.Text{String: *req.EndDate, Valid: true}
	}

	params := sqlc.CreateSubscriptionParams{
		ServiceName: req.ServiceName,
		Price:       int32(req.Price),
		UserID:      userIDPgtype,
		StartDate:   req.StartDate,
		EndDate:     endDate,
	}

	sub, err := r.queries.CreateSubscription(ctx, params)
	if err != nil {
		r.logger.Error("failed to create subscription", zap.Error(err))
		return nil, err
	}

	result := r.convertToSubscription(&sub)
	r.logger.Info("subscription created successfully", zap.String("id", result.ID.String()))
	return result, nil
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	r.logger.Info("getting subscription by id", zap.String("id", id.String()))

	idPgtype := pgtype.UUID{}
	if err := idPgtype.Scan(id.String()); err != nil {
		return nil, err
	}

	sub, err := r.queries.GetSubscription(ctx, idPgtype)
	if err != nil {
		r.logger.Error("failed to get subscription", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}

	result := r.convertToSubscription(&sub)
	r.logger.Info("subscription retrieved successfully", zap.String("id", id.String()))
	return result, nil
}

func (r *subscriptionRepository) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateSubscriptionRequest) (*domain.Subscription, error) {
	r.logger.Info("updating subscription", zap.String("id", id.String()))

	idPgtype := pgtype.UUID{}
	if err := idPgtype.Scan(id.String()); err != nil {
		return nil, err
	}

	current, err := r.queries.GetSubscription(ctx, idPgtype)
	if err != nil {
		return nil, err
	}

	serviceName := current.ServiceName
	if req.ServiceName != nil {
		serviceName = *req.ServiceName
	}

	price := current.Price
	if req.Price != nil {
		price = int32(*req.Price)
	}

	startDate := current.StartDate
	if req.StartDate != nil {
		startDate = *req.StartDate
	}

	endDate := current.EndDate
	if req.EndDate != nil {
		endDate = pgtype.Text{String: *req.EndDate, Valid: true}
	}

	params := sqlc.UpdateSubscriptionParams{
		ID:          idPgtype,
		ServiceName: serviceName,
		Price:       price,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	sub, err := r.queries.UpdateSubscription(ctx, params)
	if err != nil {
		r.logger.Error("failed to update subscription", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}

	result := r.convertToSubscription(&sub)
	r.logger.Info("subscription updated successfully", zap.String("id", id.String()))
	return result, nil
}

func (r *subscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Info("deleting subscription", zap.String("id", id.String()))

	idPgtype := pgtype.UUID{}
	if err := idPgtype.Scan(id.String()); err != nil {
		return err
	}

	if err := r.queries.DeleteSubscription(ctx, idPgtype); err != nil {
		r.logger.Error("failed to delete subscription", zap.String("id", id.String()), zap.Error(err))
		return err
	}

	r.logger.Info("subscription deleted successfully", zap.String("id", id.String()))
	return nil
}

func (r *subscriptionRepository) List(ctx context.Context, filter *ListSubscriptionsFilter) ([]*domain.Subscription, int64, error) {
	r.logger.Info("listing subscriptions",
		zap.Int("limit", filter.Limit),
		zap.Int("offset", filter.Offset),
	)

	var userID pgtype.UUID
	if filter.UserID != nil {
		if err := userID.Scan(filter.UserID.String()); err != nil {
			return nil, 0, err
		}
	}

	serviceName := ""
	if filter.ServiceName != nil {
		serviceName = *filter.ServiceName
	}

	listParams := sqlc.ListSubscriptionsParams{
		Column1: userID,
		Column2: serviceName,
		Limit:   int32(filter.Limit),
		Offset:  int32(filter.Offset),
	}

	subs, err := r.queries.ListSubscriptions(ctx, listParams)
	if err != nil {
		r.logger.Error("failed to list subscriptions", zap.Error(err))
		return nil, 0, err
	}

	countParams := sqlc.CountSubscriptionsParams{
		Column1: userID,
		Column2: serviceName,
	}

	count, err := r.queries.CountSubscriptions(ctx, countParams)
	if err != nil {
		r.logger.Error("failed to count subscriptions", zap.Error(err))
		return nil, 0, err
	}

	result := make([]*domain.Subscription, len(subs))
	for i, sub := range subs {
		result[i] = r.convertToSubscription(&sub)
	}

	r.logger.Info("subscriptions listed successfully", zap.Int("count", len(result)))
	return result, count, nil
}

func (r *subscriptionRepository) CalculateTotalCost(ctx context.Context, filter *TotalCostFilter) (int, error) {
	r.logger.Info("calculating total cost",
		zap.String("start_date", filter.StartDate),
		zap.String("end_date", filter.EndDate),
	)

	var userID pgtype.UUID
	if filter.UserID != nil {
		if err := userID.Scan(filter.UserID.String()); err != nil {
			return 0, err
		}
	}

	serviceName := ""
	if filter.ServiceName != nil {
		serviceName = *filter.ServiceName
	}

	params := sqlc.CalculateTotalCostParams{
		Column1:   userID,
		Column2:   serviceName,
		EndDate:   pgtype.Text{String: filter.StartDate, Valid: true},
		StartDate: filter.EndDate,
	}

	totalCost, err := r.queries.CalculateTotalCost(ctx, params)
	if err != nil {
		r.logger.Error("failed to calculate total cost", zap.Error(err))
		return 0, err
	}

	result := int(totalCost)
	r.logger.Info("total cost calculated successfully", zap.Int("total_cost", result))
	return result, nil
}

func (r *subscriptionRepository) convertToSubscription(sub *sqlc.Subscription) *domain.Subscription {
	userID := uuid.UUID{}
	if sub.UserID.Valid {
		copy(userID[:], sub.UserID.Bytes[:])
	}

	id := uuid.UUID{}
	if sub.ID.Valid {
		copy(id[:], sub.ID.Bytes[:])
	}

	result := &domain.Subscription{
		ID:          id,
		ServiceName: sub.ServiceName,
		Price:       int(sub.Price),
		UserID:      userID,
		StartDate:   sub.StartDate,
	}

	if sub.EndDate.Valid {
		result.EndDate = &sub.EndDate.String
	}

	if sub.CreatedAt.Valid {
		result.CreatedAt = sub.CreatedAt.Time
	}

	if sub.UpdatedAt.Valid {
		result.UpdatedAt = sub.UpdatedAt.Time
	}

	return result
}
