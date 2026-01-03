package engram

import (
	pb "engram/proto/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	pb.AgentServiceClient
	conn *grpc.ClientConn
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		AgentServiceClient: pb.NewAgentServiceClient(conn),
		conn:               conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
