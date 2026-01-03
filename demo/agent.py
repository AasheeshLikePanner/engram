import time
import grpc
import sys
from engram.v1 import agent_pb2, agent_pb2_grpc

def run_demo():
    print("🚀 Connecting to Engram Engine (localhost:50051)...")
    try:
        channel = grpc.insecure_channel('localhost:50051')
        stub = agent_pb2_grpc.AgentServiceStub(channel)
        
        # 1. Create the persistent agent
        print("\n🧠 Creating a durable agent...")
        resp = stub.CreateAgent(agent_pb2.CreateAgentRequest(
            goal="Tell a long, detailed science fiction story. Use the [CLOCK_TOOL] occasionally.",
            model="llama3.1:8b" # Change to your local model
        ))
        agent_id = resp.agent_id
        print(f"✅ Agent created: {agent_id}")
        
        # 2. Interactive Stream
        print("\n📖 Starting the story stream. KILL THE ENGINE (docker-compose down) at any time!")
        print("-" * 50)
        
        def send_message():
            yield agent_pb2.ClientMessage(
                agent_id=agent_id,
                user_input=agent_pb2.UserInput(content="Start your story now. Make it long and stop for breath occasionally.")
            )

        while True:
            try:
                for response in stub.Chat(send_message()):
                    if response.HasField('text'):
                        # Print with a slight delay to make it readable in the video
                        for char in response.text.content:
                            sys.stdout.write(char)
                            sys.stdout.flush()
                            time.sleep(0.01) 
                    elif response.HasField('tool_call'):
                        print(f"\n\n🛠️  [TOOL_CALL] Executing: {response.tool_call.name}")
                        print(f"📦 Args: {response.tool_call.arguments}")
                
                print("\n\n✅ Stream finished naturally.")
                break
            except grpc.RpcError as e:
                print(f"\n\n⚠️  CONNECTION LOST: {e.code()}")
                print("🔄 Waiting for engine to restart... (Run 'docker-compose up' now!)")
                while True:
                    try:
                        time.sleep(2)
                        # Heartbeat check
                        stub.GetAgent(agent_pb2.GetAgentRequest(agent_id=agent_id))
                        print("✨ Engine is back! Resuming soul recovery...")
                        break
                    except:
                        pass

    except Exception as e:
        print(f"❌ Error: {e}")

if __name__ == "__main__":
    run_demo()
