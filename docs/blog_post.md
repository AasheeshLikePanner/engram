# The End of Amnesia: Why We Built Engram 🧠

**Imagine hiring a brilliant employee who forgets everything every time they go to sleep.**

They solve complex problems, write amazing code, and understand your business perfectly. But the moment they close their eyes (or their process restarts), it's all gone. Day 2 is exactly like Day 1. You have to re-explain *everything*.

This is the reality of AI Agents today. We build them as **scripts**—ephemeral, fragile, and prone to "dying" whenever a container crashes or an API times out.

We built **Engram** to change that.

## ⚡ What is Engram?

Engram is a **high-performance, durable runtime** for AI agents. It does one thing, and it does it perfectly: it gives your agents **indestructible memory**.

It decouples the *Compute* (the LLM reasoning) from the *State* (the memory). Even if you kill the entire server, an Engram agent simply pauses. When you bring it back, it picks up exactly where it left off, down to the last token.

## 🛠️ How to Use It

We designed Engram to be invisible. You don't need to learn a complex framework. Just run the engine and use the SDK.

### 1. The Setup
The engine is a single binary (packaged in Docker) written in **Go**. It handles the heavy lifting of state management and event streaming.

```bash
docker run -d -p 50051:50051 aasheeshlikepanner/engram-engine:v0.1.0
pip install engram-sdk
```

### 2. The Code
Here is how simple it is to create a persistent agent in Python:

```python
import grpc
from engram.v1 import agent_pb2, agent_pb2_grpc

# Connect to the local Engram Engine
channel = grpc.insecure_channel('localhost:50051')
stub = agent_pb2_grpc.AgentServiceStub(channel)

# Create a specialized agent
resp = stub.CreateAgent(agent_pb2.CreateAgentRequest(
    goal="Optimize the backend database queries.",
    model="llama3.1:70b",
    system_prompt="You are a senior DBA using Postgres..."
))
agent_id = resp.agent_id

# Chat interactively (Context is saved automatically!)
user_msg = agent_pb2.ClientMessage(
    agent_id=agent_id,
    user_input=agent_pb2.UserInput(content="We are seeing slow joins on the 'users' table.")
)

for msg in stub.Chat(iter([user_msg])):
    if msg.HasField('text'):
        print(f"Agent: {msg.text.content}")
```

If you crash this script right now and run it again tomorrow, asking *"What table were we talking about?"*, the agent will reply: *"The 'users' table."*

## 🏎️ Why It’s So Fast

Speed isn't just a feature; it's a requirement for modern interaction. We didn't wrap a Python library in a REST API. We built a **native runtime in Go**.

1.  **gRPC Streaming**: Instead of clunky HTTP requests, we use bidirectional streaming. This gives us **sub-millisecond latency**.
2.  **Zero-Overhead Memory**: State is managed asynchronously. Writing to the event log doesn't block the generation of the next token.
3.  **The Numbers**:
    - **Agent Creation**: ~4ms
    - **Time-to-First-Token**: ~0.7ms

This means Engram feels *instant*. It adds virtually 0ms latency to your LLM calls.

## 🌍 Shaping the Future

Why does this matter?

We are moving from an era of **Chatbots** to **Digital Colleagues**.
- **Chatbots** are disposable. You talk, you get an answer, you leave.
- **Colleagues** share a history. They learn your preferences, remember past projects, and grow with you.

For that future to exist, durability cannot be an afterthought. It has to be the **kernel**.

Engram provides that kernel. It enables a future where agents run for **months, not minutes**. Where an AI developer can onboard on Monday, work on a codebase for a week, sleep for the weekend, and present their PR on Monday—without forgetting a single line of code.

Stop building scripts. Start building lighthouses.

**[Get Started on GitHub](https://github.com/zynta/engram-engine)**
