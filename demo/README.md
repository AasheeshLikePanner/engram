# Engram Live Demo 🚀

This folder contains the official scripts for the Engram durability and sandboxing demo.

## 🏃 Instructions

### 1. Start the Environment
```bash
docker-compose up --build
```

### 2. Run the Agent
In a separate terminal:
```bash
python3 demo/agent.py
```

### 3. The "Crash" Moment
While the agent is printing text:
1. Go back to the first terminal and hit `Ctrl+C`.
2. Wait for `agent.py` to show the "RECOVERY" message.
3. Run `docker-compose up` again.
4. Watch the agent pick up exactly where it left off.

## 🛠️ Tools
- `tools/clock.wasm`: A simple Go-based WASM tool that is automatically mounted into the engine sandbox.
