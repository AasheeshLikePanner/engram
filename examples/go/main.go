package main

import (
	"context"
	engram "engram/pkg/sdk/go"
	pb "engram/proto/v1"
	"fmt"
	"log"
)

func main() {
	client, err := engram.NewClient("localhost:50051")
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	resp, err := client.CreateAgent(context.Background(), &pb.CreateAgentRequest{
		Goal:  "Hello World Agent",
		Model: "llama3.1",
	})
	if err != nil {
		log.Fatalf("failed to create agent: %v", err)
	}

	fmt.Printf("Created Agent: %s\n", resp.AgentId)
}
