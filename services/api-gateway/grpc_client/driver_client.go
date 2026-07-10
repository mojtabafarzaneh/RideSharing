package grpc_client

import (
	pb "ride-sharing/shared/proto/driver"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type driverServiceClient struct {
	Client pb.DriverServiceClient
	conn   *grpc.ClientConn
}

func NewDriverServiceClient() (*driverServiceClient, error) {
	driverServiceURL := "driver-service:9092"
	//tripServiceURL := "localhost:9083"
	conn, err := grpc.NewClient(driverServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	client := pb.NewDriverServiceClient(conn)

	return &driverServiceClient{
		Client: client,
		conn:   conn,
	}, nil

}

func (c *driverServiceClient) Close() {

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return
		}
	}
}
