package repository

import (
	"context"
	"fmt"
	"ride-sharing/services/trip-service/internal/domain"
)

type inmemRepository struct {
	trips     map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel
}

func NewInmemRepository() domain.TripRepository {
	return &inmemRepository{
		trips:     make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}
}

func (i *inmemRepository) SaveTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	i.trips[trip.ID.Hex()] = trip
	return trip, nil
}

func (i *inmemRepository) SaveRideFare(ctx context.Context, fare *domain.RideFareModel) error {

	i.rideFares[fare.ID.Hex()] = fare

	return nil
}

func (i *inmemRepository) GetRideFareByID(ctx context.Context, id string) (*domain.RideFareModel, error) {

	res, exist := i.rideFares[id]

	if !exist {
		return nil, fmt.Errorf("could not find any fare with the provided id")
	}

	return res, nil

}
