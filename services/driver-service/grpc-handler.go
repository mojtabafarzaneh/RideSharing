package main

import (
	"context"
	"log"
	pb "ride-sharing/shared/proto/driver"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcHandler struct {
	pb.UnimplementedDriverServiceServer

	Service *Service
}

func NewGrpcHandler(s *grpc.Server, service *Service) *grpcHandler {

	handler := &grpcHandler{
		Service: service,
	}
	pb.RegisterDriverServiceServer(s, handler)
	return handler

}

func (h *grpcHandler) RegisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	driverId := req.GetDriverID()
	packageSlug := req.GetPackageSlug()

	register, err := h.Service.RegisterDriver(driverId, packageSlug)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to register driver: %v", err)

	}

	return &pb.RegisterDriverResponse{
		Driver: register,
	}, nil
}
func (h *grpcHandler) UnregisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	driverId := req.GetDriverID()

	h.Service.UnregisterDriver(driverId)

	return &pb.RegisterDriverResponse{
		Driver: &pb.Driver{
			Id: driverId,
		},
	}, nil
}
