import grpc
import time
import sys
from engram.v1 import agent_pb2
from engram.v1 import agent_pb2_grpc

def chat_with_agent(stub, agent_id, text):
    msg = agent_pb2.ClientMessage(
        agent_id=agent_id,
        user_input=agent_pb2.UserInput(content=text)
    )
    responses = stub.Chat(iter([msg]))
    for r in responses:
        if r.HasField('text'):
            return r.text.content
        if r.HasField('tool_call'):
            return f"[TOOL_CALL] {r.tool_call.tool_name}"
        if r.HasField('tool_result'):
            return f"[TOOL_RESULT] {r.tool_result.output}"
    return ""

def run_test():
    channel = grpc.insecure_channel('localhost:50051')
    stub = agent_pb2_grpc.AgentServiceStub(channel)

    print("--- Phase 1: Establish Temporal Context & Sandbox ---")
    resp = stub.CreateAgent(agent_pb2.CreateAgentRequest(
        goal="Verify time persistence and sandbox",
        model="mock-model"
    ))
    agent_id = resp.agent_id
    print(f"Created Agent: {agent_id}")

    # Set the 'remembered' time
    context_msg = "It is exactly 3:20 PM right now. Remember this."
    print(f"Sending: {context_msg}")
    reply = chat_with_agent(stub, agent_id, context_msg)
    print(f"Agent: {reply}")

    print("\n--- Phase 2: Test Sandbox Execution (Clock Tool) ---")
    tool_trigger = "[CLOCK_TOOL] check environment time"
    print(f"Sending: {tool_trigger}")
    # This should trigger [TOOL_CALL] from MockLLM
    reply = chat_with_agent(stub, agent_id, tool_trigger)
    print(f"Agent: {reply}")
    
    # Wait for worker to run the tool and generate a second LLM response
    # The SDK 'Chat' returns a stream, we might need to wait for the result
    # For this test, we'll just check if we got a tool result event in the stream
    print("Waiting for tool result...")
    time.sleep(2)

    print("\n--- Phase 3: Verify Recovery of Temporal Context ---")
    query = "What time did I say it was earlier?"
    print(f"Querying: {query}")
    reply = chat_with_agent(stub, agent_id, query)
    print(f"Agent: {reply}")

    if "3:20 PM" in reply:
        print("\n✅ SUCCESS: Agent recovered the specific time context!")
    else:
        print("\n❌ FAILED: Agent lost the time context.")

    channel.close()

if __name__ == "__main__":
    run_test()
