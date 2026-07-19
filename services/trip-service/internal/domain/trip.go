package domain

import (
	"context"
	tripType "ride-sharing/services/trip-service/pkg/types"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

	pbd "ride-sharing/shared/proto/driver"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripModel struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	UserID   string             `bson:"user_id,omitempty"`
	Status   string             `bson:"status,omitempty"`
	RideFare *RideFareModel     `bson:"ride_fare,omitempty"`
	Driver   *pb.TripDriver
}

func (t *TripModel) ToPorto() *pb.Trip {
	return &pb.Trip{
		Id:           t.ID.Hex(),
		UserID:       t.UserID,
		SelectedFare: t.RideFare.ToProto(),
		Status:       t.Status,
		Driver:       t.Driver,
		Route:        t.RideFare.Route.ToProto(),
	}
}

type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
	GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripType.OsrmApiResponse, error)
	EstimatePackagesPriceWithRoute(route *tripType.OsrmApiResponse) []*RideFareModel
	GenerateTripFares(ctx context.Context, fares []*RideFareModel, userID string, route *tripType.OsrmApiResponse) ([]*RideFareModel, error)

	GetAndValidateFare(ctx context.Context, fareId, userId string) (*RideFareModel, error)
	GetTripById(ctx context.Context, tripId string) (*TripModel, error)
	UpdateTrip(ctx context.Context, tripId string, status string, driver *pbd.Driver) error
}
type TripRepository interface {
	SaveRideFare(ctx context.Context, fare *RideFareModel) error
	SaveTrip(ctx context.Context, trip *TripModel) (*TripModel, error)
	GetRideFareByID(ctx context.Context, id string) (*RideFareModel, error)
	GetTripById(ctx context.Context, tripId string) (*TripModel, error)
	UpdateTrip(ctx context.Context, tripId string, status string, driver *pbd.Driver) error
}
