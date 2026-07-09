package main

import pb "ride-sharing/shared/proto/driver"

type Service struct {
	drivers []*driverInMap
}

type driverInMap struct {
	Driver *pb.Driver
}

func newService() *Service {
	return &Service{
		drivers: make([]*driverInMap, 0),
	}
}
