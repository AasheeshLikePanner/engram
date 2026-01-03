import grpc
from engram.v1 import agent_pb2
from engram.v1 import agent_pb2_grpc

def run():
    with grpc.insecure_channel('localhost:50051') as channel:
        stub = agent_pb2_grpc.AgentServiceStub(channel)
        
        # Create an agent
        response = stub.CreateAgent(agent_pb2.CreateAgentRequest(
            goal="Tell me a joke",
            model="llama3.1"
        ))
        print(f"Created Agent ID: {response.agent_id}")

        # List agents
        agents = stub.ListAgents(agent_pb2.ListAgentsRequest())
        print(f"Found {len(agents.agents)} agents")

if __name__ == "__main__":
    run()
