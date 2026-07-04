package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripModel struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	UserID   string             `bson:"user_id,omitempty"`
	Status   string             `bson:"status,omitempty"`
	RideFare *RideFareModel     `bson:"ride_fare,omitempty"`
}
