import grpc
import time
import sys
from engram.v1 import agent_pb2, agent_pb2_grpc

def test_sandbox():
    print("🧪 Testing WASM Sandbox Integration...")
    channel = grpc.insecure_channel('localhost:50051')
    stub = agent_pb2_grpc.AgentServiceStub(channel)
    
    resp = stub.CreateAgent(agent_pb2.CreateAgentRequest(
        goal="Test the clock tool.",
        model="mock"
    ))
    agent_id = resp.agent_id
    
    # MockLLM triggers tool call if user message contains [CLOCK_TOOL]
    def send_message():
        yield agent_pb2.ClientMessage(
            agent_id=agent_id,
            user_input=agent_pb2.UserInput(content="Please use the [CLOCK_TOOL] now.")
        )

    print(f"📡 Requesting clock tool for agent {agent_id}...")
    for response in stub.Chat(send_message()):
        if response.HasField('tool_call'):
            print(f"✅ Received tool call: {response.tool_call.tool_name}")
            print(f"📦 Arguments: {response.tool_call.input_json}")
        elif response.HasField('text'):
            print(f"🤖 Agent: {response.text.content}")
        
    print("\n🏁 Sandbox test complete.")

if __name__ == "__main__":
    test_sandbox()
