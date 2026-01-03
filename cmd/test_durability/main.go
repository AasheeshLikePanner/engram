package main

import (
	"context"
	pb "engram/proto/v1"
	"fmt"
	"log"
	"time"

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

	// 1. Create Agent
	fmt.Println("--- Step 1: Create Agent ---")
	createRes, err := c.CreateAgent(ctx, &pb.CreateAgentRequest{
		Goal:  "Test complex scenario with tools and edits",
		Model: "llama3",
	})
	if err != nil {
		log.Fatalf("could not create agent: %v", err)
	}
	agentID := createRes.AgentId
	fmt.Printf("Created Agent: %s\n", agentID)

	// 2. Start Chat and Stream
	fmt.Println("\n--- Step 2: Start Chat Stream ---")
	stream, err := c.Chat(ctx)
	if err != nil {
		log.Fatalf("could not start chat: %v", err)
	}

	// Send initial message
	err = stream.Send(&pb.ClientMessage{
		AgentId: agentID,
		Payload: &pb.ClientMessage_UserInput{
			UserInput: &pb.UserInput{Content: "Please execute a [TOOL_CALL] now."},
		},
	})
	if err != nil {
		log.Fatalf("could not send message: %v", err)
	}

	// Listen for events
	done := make(chan bool)
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				fmt.Printf("Stream closed or error: %v\n", err)
				return
			}
			fmt.Printf("<<< Received Event: Seq=%d, Payload=%T\n", msg.SequenceId, msg.Payload)

			// If we get a tool call, we signal to pause/edit
			if _, ok := msg.Payload.(*pb.ServerMessage_ToolCall); ok {
				fmt.Println("\n--- Step 3: Tool Call Detected! Pausing for HITL ---")
				done <- true
				return
			}
		}
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		log.Fatal("Timeout waiting for tool call")
	}

	// 4. Send Edit and Resume
	fmt.Println("\n--- Step 4: Sending Edit Intervention ---")
	err = stream.Send(&pb.ClientMessage{
		AgentId: agentID,
		Payload: &pb.ClientMessage_Edit{
			Edit: &pb.Edit{
				Description: "Change the tool input to something else",
			},
		},
	})
	if err != nil {
		log.Fatalf("could not send edit: %v", err)
	}

	fmt.Println("\n--- Step 5: Resuming Agent ---")
	_, err = c.ControlAgent(ctx, &pb.ControlAgentRequest{
		AgentId: agentID,
		Command: &pb.ControlAgentRequest_Resume{Resume: &pb.Resume{}},
	})
	if err != nil {
		log.Fatalf("could not resume agent: %v", err)
	}

	// Wait for the result
	fmt.Println("\n--- Step 6: Waiting for Tool Result and Final Answer ---")
	for i := 0; i < 5; i++ {
		msg, err := stream.Recv()
		if err != nil {
			break
		}
		fmt.Printf("<<< Received Event: Seq=%d, Payload=%T\n", msg.SequenceId, msg.Payload)
		if _, ok := msg.Payload.(*pb.ServerMessage_Text); ok {
			fmt.Println("SUCCESS: Received final response after tool execution!")
			break
		}
	}
}
