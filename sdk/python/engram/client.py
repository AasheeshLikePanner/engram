import grpc
from typing import Optional, List, Iterator
from engram.v1 import agent_pb2
from engram.v1 import agent_pb2_grpc

class EngramClient:
    """
    High-level Python SDK for the Engram Agent Engine.
    Handles authentication and provides a clean API for interacting with agents.
    """
    def __init__(self, target: str = 'localhost:50051', api_key: Optional[str] = None):
        self.target = target
        self.api_key = api_key
        self.channel = grpc.insecure_channel(target)
        self.stub = agent_pb2_grpc.AgentServiceStub(self.channel)

    def _get_metadata(self):
        if self.api_key:
            return [('x-api-key', self.api_key)]
        return []

    def create_agent(self, goal: str, model: str, metadata: Optional[dict] = None) -> agent_pb2.Agent:
        request = agent_pb2.CreateAgentRequest(
            goal=goal,
            model=model,
            metadata=metadata or {}
        )
        response = self.stub.CreateAgent(request, metadata=self._get_metadata())
        return response.agent

    def get_agent(self, agent_id: str) -> agent_pb2.Agent:
        request = agent_pb2.GetAgentRequest(agent_id=agent_id)
        return self.stub.GetAgent(request, metadata=self._get_metadata())

    def list_agents(self, status_filter: str = "") -> List[agent_pb2.Agent]:
        request = agent_pb2.ListAgentsRequest(status_filter=status_filter)
        response = self.stub.ListAgents(request, metadata=self._get_metadata())
        return list(response.agents)

    def delete_agent(self, agent_id: str):
        request = agent_pb2.DeleteAgentRequest(agent_id=agent_id)
        self.stub.DeleteAgent(request, metadata=self._get_metadata())

    def chat(self, agent_id: str, message: str) -> Iterator[agent_pb2.ServerMessage]:
        """
        Stream chat messages to and from an agent.
        """
        def request_generator():
            # First message identifies the agent and provides input
            # In a real bidirectional stream, we might want to keep this open
            yield agent_pb2.ClientMessage(
                agent_id=agent_id,
                user_input=agent_pb2.UserInput(content=message)
            )

        # We return the stream directly
        return self.stub.Chat(request_generator(), metadata=self._get_metadata())

    def control_agent(self, agent_id: str, command: str) -> str:
        """
        Send a control command: pause, resume, cancel, or retry.
        """
        req_kwargs = {'agent_id': agent_id}
        if command == "pause":
            req_kwargs['pause'] = agent_pb2.Pause()
        elif command == "resume":
            req_kwargs['resume'] = agent_pb2.Resume()
        elif command == "cancel":
            req_kwargs['cancel'] = agent_pb2.Cancel()
        elif command == "retry":
            req_kwargs['retry'] = agent_pb2.Retry()
        else:
            raise ValueError(f"Unknown command: {command}")

        request = agent_pb2.ControlAgentRequest(**req_kwargs)
        response = self.stub.ControlAgent(request, metadata=self._get_metadata())
        return response.status

    def close(self):
        self.channel.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()
