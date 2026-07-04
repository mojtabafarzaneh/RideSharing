package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	UserID            string             `bson:"user_id,omitempty"`
	PackageSlug       string             `bson:"package_slug,omitempty"`
	TotalPriceInCents int64              `bson:"total_price_in_cents,omitempty"`
	ExpiresAt         primitive.DateTime `bson:"expires_at,omitempty"`
}

type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
}
type TripRepository interface {
	SaveTrip(ctx context.Context, trip *TripModel) (TripModel, error)
}
