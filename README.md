# Engram 🧠

> **The Durable AI Agent Engine.**
> High-performance, event-sourced runtime for persistent AI agents.

[![Python SDK](https://img.shields.io/pypi/v/engram-sdk)](https://pypi.org/project/engram-sdk/)
[![Docker Image](https://img.shields.io/docker/v/zynta/engram-engine)](https://hub.docker.com/r/zynta/engram-engine)

Engram is a **Stateful AI Engine** designed to solve the "amnesia problem" of modern LLM agents. Unlike ephemeral scripts that lose context when they crash, Engram treats agents as **persistent processes**. Every thought, tool call, and observation is essentially stored in an event log, allowing agents to **pause, crash, restart, and recover** exactly where they left off.

## ✨ Why Engram?

- **🛡️ Indestructible Context**: Built on an Event Sourcing architecture. Kill the container, restart it, and your agent remembers *everything*.
- **⚡ Sub-Millisecond Latency**: Written in **Go** with high-concurrency gRPC streaming.
    - Agent Creation: **~4ms**
    - Time-to-First-Token: **~0.7ms**
- **🔌 Pluggable Storage**: Runs on **Redis** (Production) or **BadgerDB** (Embedded/Local).
- **🐍 Native Python SDK**: Interact with your agents using a clean, typed Python client.

---

## 🚀 Quick Start

### 1. Run the Engine
Use the pre-built Docker image. You need a Redis instance for persistence.

```bash
# Run Redis and Engram Engine
docker run -d --name redis -p 6379:6379 redis:alpine

docker run -d --name engram \
  -p 50051:50051 \
  -e STORE_TYPE=redis \
  -e REDIS_URL=host.docker.internal:6379 \
  aasheeshlikepanner/engram-engine:v0.1.0
```

### 2. Install the SDK
```bash
pip install engram-sdk
```

### 3. Create & Chat with an Agent
```python
import grpc
from engram.v1 import agent_pb2, agent_pb2_grpc

# Connect to the engine
channel = grpc.insecure_channel('localhost:50051')
stub = agent_pb2_grpc.AgentServiceStub(channel)

# Create an agent
resp = stub.CreateAgent(agent_pb2.CreateAgentRequest(
    goal="Help the user verify durability.",
    model="llama3.1:70b"
))
agent_id = resp.agent_id
print(f"Created Agent: {agent_id}")

# Chat interactively
user_msg = agent_pb2.ClientMessage(
    agent_id=agent_id,
    user_input=agent_pb2.UserInput(content="Remember that the secret code is 42.")
)

# Send message and stream response
for response in stub.Chat(iter([user_msg])):
    if response.HasField('text'):
        print(f"Agent: {response.text.content}")
```

---

## 🏗️ Architecture

Engram decouples the **Compute** (LLM/Logic) from the **State** (Memory).

1.  **Event Log**: The source of truth. Every interaction is appended to an immutable log (Redis Stream).
2.  **Worker**: A reactive process that subscribes to the log. It reads the specific history, feeds it to the LLM, and appends the new response as a new event.
3.  **Snapshotting**: (Coming Soon) Periodic snapshots to squash history and speed up context loading.

This means you can upgrade the engine functionality *without* touching the agent's memory.

## 📦 Installation

**Docker**
```bash
docker pull aasheeshlikepanner/engram-engine:latest
```

**Python SDK**
```bash
pip install engram-sdk
```

## 📜 License
MIT License. Built with ❤️ for the AI Engineering community.
