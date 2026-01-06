from engram import EngramClient
import time

def run():
    # Initialize the client (assumes engine is running on localhost:50051)
    # If you set an API_KEY in your engine config, pass it here.
    client = EngramClient(target='localhost:50051', api_key='secret-token')
    
    try:
        print("🚀 Creating a new durable agent...")
        agent = client.create_agent(
            goal="Tell me a joke and then check the current time.",
            model="llama-3.3-70b-versatile"
        )
        print(f"✅ Agent Created! ID: {agent.agent_id}")

        print("\n💬 Sending message...")
        stream = client.chat(agent.agent_id, "Go ahead!")

        for msg in stream:
            if msg.HasField('thought'):
                print(f"💭 Thought: {msg.thought.content}")
            elif msg.HasField('tool_call'):
                print(f"🛠️  Tool Call: {msg.tool_call.tool_name}")
            elif msg.HasField('text'):
                print(f"🤖 Response: {msg.text.content}")
            elif msg.HasField('status'):
                print(f"ℹ️  Status: {msg.status.message}")

    except Exception as e:
        print(f"❌ Error: {e}")
    finally:
        client.close()

if __name__ == "__main__":
    run()