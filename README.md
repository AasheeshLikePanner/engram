# Engram 🧠

> **The Durable AI Agent Engine.**
> High-performance, event-sourced runtime for persistent AI agents.

[![Python SDK](https://img.shields.io/pypi/v/engram-sdk)](https://pypi.org/project/engram-sdk/0.1.1/)
[![Docker Image](https://img.shields.io/docker/v/aasheeshlikepanner/engram-engine)](https://hub.docker.com/r/aasheeshlikepanner/engram-engine)

Engram is a **Stateful AI Engine** designed to solve the "amnesia problem" of modern LLM agents. Unlike ephemeral scripts that lose context when they crash, Engram treats agents as **persistent processes**. Every thought, tool call, and observation is essentially stored in an event log, allowing agents to **pause, crash, restart, and recover** exactly where they left off.

## ✨ Why Engram?

- **🛡️ Indestructible Context**: Built on an Event Sourcing architecture. Kill the container, restart it, and your agent remembers *everything* through a deterministic replay of its event log.
- **🏗️ Secure WASM Sandboxing**: Agents run tools in a high-isolation **WASM sandbox** (powered by `wazero`). This ensures CPU/Memory isolation and prevents unauthenticated filesystem or network access.
- **🕰️ Temporal Precision**: Replayed history includes original event timestamps, while the agent remains aware of the *current* system time. This solves the "time-drift" problem in recovered agents.
- **⚡ Sub-Millisecond Latency**: Written in **Go** with high-concurrency gRPC streaming.
    - Agent Creation: **~4ms** | TTFT: **~0.7ms**
- **🔌 Pluggable Storage**: Runs on **Redis Streams** (Scalable Production) or **BadgerDB** (Embedded/Local).
- **🐍 Native Python SDK**: Fully typed and asynchronous-ready client.

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
  aasheeshlikepanner/engram-engine:v0.1.1
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

## 🏗️ Technical Pillars

### 1. Event Sourced Souls
Engram decouples the **Compute** (LLM/Logic) from the **State** (Memory).
- **Immutable Log**: Every interaction is appended to an immutable log (Redis Stream).
- **Reactive Workers**: Workers subscribe to updates, replaying the log to reconstruct context before every step.
- **Temporal Recovery**: Replays include `[Timestamp]` markers, ensuring agents understand the chronological order of events even after a crash.

### 2. WASM "Level 0" Sandboxing 🛡️
Security isn't an afterthought; it's the foundation.
- **Dynamic Tools**: Place any `.wasm` binary in the `tools/` directory. The engine resolves and executes them at runtime.
- **Hard Isolation**: Tools run in a strictly bounded WASM runtime (powered by `wazero`) with no access to the host's OS-level primitives.
- **Idempotent Results**: Every sandbox output is cached in the event log, preventing double-execution during recovery loops.

### 3. High-Performance gRPC
- **Bidirectional Streaming**: Real-time interaction with negligible overhead.
- **Cross-Language Support**: Native SDKs for Go and Python out of the box.

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
