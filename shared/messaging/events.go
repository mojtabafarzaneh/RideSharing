package messaging

import (
	pb "ride-sharing/shared/proto/trip"
)

const (
	FindAvailableDriversQueue = "find_available_drivers"
	DriverCMDTripRequestQueue = "driver_cmd_trip_request"
)

type TripEventData struct {
	Trip *pb.Trip `json:"trip"`
}
