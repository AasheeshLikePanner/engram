import sys
import os
import time
from engram import EngramClient

def interactive_demo():
    print("="*50)
    print("      ENGRAM: Durable AI Agent Engine Demo")
    print("="*50)
    
    # 1. Setup
    api_key = os.getenv("API_KEY", "secret-token")
    target = os.getenv("ENGINE_URL", "localhost:50051")
    client = EngramClient(target=target, api_key=api_key)
    
    try:
        # 2. Agent Configuration
        print("\n[1/3] Initializing Agent...")
        goal = "You are a helpful research assistant with access to a clock tool."
        model = "llama-3.3-70b-versatile"
        
        agent = client.create_agent(goal=goal, model=model)
        agent_id = agent.agent_id
        print(f"✨ Agent '{agent_id}' is now live and durable.")

        # 3. Interactive Loop
        print("\n[2/3] Entering Interactive Mode (type 'exit' to quit)")
        while True:
            user_input = input("\n👤 You: ")
            if user_input.lower() in ['exit', 'quit']:
                break

            print("\n🤖 Assistant is thinking...")
            responses = client.chat(agent_id, user_input)
            
            for msg in responses:
                if msg.HasField('thought'):
                    # Using dim/grey style for thoughts if possible, otherwise plain
                    print(f"  💭 {msg.thought.content}")
                
                elif msg.HasField('tool_call'):
                    print(f"  🛠️  Executing tool: {msg.tool_call.tool_name}...")
                
                elif msg.HasField('tool_result'):
                    print(f"  📦 Tool returned: {msg.tool_result.output}")
                
                elif msg.HasField('text'):
                    print(f"\n🤖: {msg.text.content}")
                
                elif msg.HasField('error'):
                    print(f"\n❌ System Error: {msg.error.message}")

        # 4. Persistence Showcase
        print("\n[3/3] Persistence Check...")
        print("Even if the connection closed, the agent's state is safe in the store.")
        agents = client.list_agents()
        print(f"Total agents in store: {len(agents)}")
        
    except KeyboardInterrupt:
        print("\n\nGoodbye!")
    except Exception as e:
        print(f"\n💥 Demo Error: {e}")
    finally:
        client.close()

if __name__ == "__main__":
    interactive_demo()