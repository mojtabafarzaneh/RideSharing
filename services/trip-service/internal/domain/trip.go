package domain

import (
	"context"
	tripType "ride-sharing/services/trip-service/pkg/types"
	"ride-sharing/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripModel struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	UserID   string             `bson:"user_id,omitempty"`
	Status   string             `bson:"status,omitempty"`
	RideFare *RideFareModel     `bson:"ride_fare,omitempty"`
}

type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
	GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripType.OsrmApiResponse, error)
}
type TripRepository interface {
	SaveTrip(ctx context.Context, trip *TripModel) (*TripModel, error)
}
