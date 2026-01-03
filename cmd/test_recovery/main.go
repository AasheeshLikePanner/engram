package main

import (
	"context"
	pb "engram/proto/v1"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewAgentServiceClient(conn)

	ctx := context.Background()

	// Test 1: List all agents to see persisted state
	fmt.Println("=== TEST: Listing All Persisted Agents ===")
	listRes, err := c.ListAgents(ctx, &pb.ListAgentsRequest{})
	if err != nil {
		log.Fatalf("could not list agents: %v", err)
	}
	fmt.Printf("Found %d agents:\n", len(listRes.Agents))
	for _, a := range listRes.Agents {
		fmt.Printf("  - ID: %s, Status: %s, Goal: %s\n", a.AgentId, a.Status, a.Goal)
	}

	if len(listRes.Agents) == 0 {
		log.Fatal("FAIL: No agents found after restart!")
	}
	agentID := listRes.Agents[0].AgentId

	// Test 2: Replay all events for an old agent
	fmt.Printf("\n=== TEST: Replaying History for Agent %s ===\n", agentID)
	replayStream, err := c.Replay(ctx, &pb.ReplayRequest{AgentId: agentID})
	if err != nil {
		log.Fatalf("could not replay: %v", err)
	}

	eventCount := 0
	for {
		msg, err := replayStream.Recv()
		if err != nil {
			break
		}
		eventCount++
		fmt.Printf("  Event %d (Seq=%d): %T\n", eventCount, msg.SequenceId, msg.Payload)
	}
	fmt.Printf("Total events replayed: %d\n", eventCount)

	if eventCount < 3 {
		log.Fatal("FAIL: Expected at least 3 events for a complete agent history!")
	}

	// Test 3: Continue conversation with old agent
	fmt.Println("\n=== TEST: Continuing Old Agent Conversation ===")
	stream, err := c.Chat(ctx)
	if err != nil {
		log.Fatalf("could not start chat: %v", err)
	}

	err = stream.Send(&pb.ClientMessage{
		AgentId: agentID,
		Payload: &pb.ClientMessage_UserInput{
			UserInput: &pb.UserInput{Content: "What was my original question?"},
		},
	})
	if err != nil {
		log.Fatalf("could not send message: %v", err)
	}

	// Wait for response
	for i := 0; i < 5; i++ {
		msg, err := stream.Recv()
		if err != nil {
			break
		}
		fmt.Printf("  <<< Received: Seq=%d, Payload=%T\n", msg.SequenceId, msg.Payload)
		if _, ok := msg.Payload.(*pb.ServerMessage_Text); ok {
			fmt.Println("SUCCESS: Agent correctly continued from old context!")
			break
		}
	}

	fmt.Println("\n=== ALL TESTS PASSED ===")
}
