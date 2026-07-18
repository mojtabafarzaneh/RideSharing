package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/pkg/types"
	"ride-sharing/shared/proto/trip"
	sharedTypes "ride-sharing/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type service struct {
	repo domain.TripRepository
}

func NewService(repo domain.TripRepository) *service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {

	t := &domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserID:   fare.UserID,
		Status:   "pending",
		RideFare: fare,
		Driver:   &trip.TripDriver{},
	}

	return s.repo.SaveTrip(ctx, t)
}

func (s *service) GetRoute(ctx context.Context, pickup, destination *sharedTypes.Coordinate) (*types.OsrmApiResponse, error) {

	url := fmt.Sprintf(
		"http://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		pickup.Longitude,
		pickup.Latitude,
		destination.Longitude,
		destination.Latitude)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get route: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response: %v", err)

	}

	var routeResp types.OsrmApiResponse

	if err := json.Unmarshal(body, &routeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the response: %v", err)
	}

	return &routeResp, nil
}

func (s *service) EstimatePackagesPriceWithRoute(route *types.OsrmApiResponse) []*domain.RideFareModel {
	baseFares := getBaseFares()

	estimatedFares := make([]*domain.RideFareModel, len(baseFares))

	for i, f := range baseFares {
		estimatedFares[i] = estimateFareRoute(f, route)
	}

	return estimatedFares
}

func (s *service) GenerateTripFares(ctx context.Context, rideFares []*domain.RideFareModel, userID string, route *types.OsrmApiResponse) ([]*domain.RideFareModel, error) {
	fares := make([]*domain.RideFareModel, len(rideFares))

	for i, f := range rideFares {
		id := primitive.NewObjectID()

		fare := &domain.RideFareModel{
			UserID:            userID,
			ID:                id,
			TotalPriceInCents: f.TotalPriceInCents,
			PackageSlug:       f.PackageSlug,
			Route:             route,
		}
		if err := s.repo.SaveRideFare(ctx, fare); err != nil {
			return nil, fmt.Errorf("failed to save fare")
		}

		fares[i] = fare
	}
	return fares, nil
}

func (s *service) GetAndValidateFare(ctx context.Context, fareId, userId string) (*domain.RideFareModel, error) {

	fare, err := s.repo.GetRideFareByID(ctx, fareId)
	if err != nil {
		return nil, fmt.Errorf("failed to get trip fare: %w", err)
	}

	if userId != fare.UserID {
		return nil, fmt.Errorf("you could not access this fare")
	}

	return fare, nil
}

func estimateFareRoute(fares *domain.RideFareModel, route *types.OsrmApiResponse) *domain.RideFareModel {
	pricingCng := types.DefaultPricingConfig()

	carPackagePrice := fares.TotalPriceInCents

	distanceKm := route.Routes[0].Distance

	durationMinutes := route.Routes[0].Duration

	distanceFare := distanceKm * pricingCng.PricePerUnitOfDistance
	durationFare := durationMinutes * pricingCng.PricingPerMinute

	totalPrice := distanceFare + durationFare + carPackagePrice

	return &domain.RideFareModel{
		TotalPriceInCents: totalPrice,
		PackageSlug:       fares.PackageSlug,
	}
}

func getBaseFares() []*domain.RideFareModel {
	return []*domain.RideFareModel{
		{
			PackageSlug:       "suv",
			TotalPriceInCents: 350,
		},
		{
			PackageSlug:       "sedan",
			TotalPriceInCents: 750,
		},
		{
			PackageSlug:       "van",
			TotalPriceInCents: 1000,
		},
		{
			PackageSlug:       "luxury",
			TotalPriceInCents: 1500,
		},
	}
}
