# Engram: "The Save-Game for AI" — Demo Script 🎥

This demo is designed to show **Engram's** unique "indestructibility" and its secure **WASM Sandboxing**.

## 🏗️ Setup Requirements
- Docker and Docker Compose installed.
- Python SDK installed (`pip install engram-sdk`).
- Two terminal windows open.

---

## Part 1: The "Amnesia" Problem (Spoken)
*"Hello! I'm Aasheesh, and today I'm showing you Engram. Most AI agents today are fragile. If your server blips or a process crashes mid-thought, the agent is gone. Context is vaporized. We're fixing that."*

---

## Part 2: Running the Engine (Visual)
**Step 1:** Show a clean `docker-compose.yaml`.
**Step 2:** Run `docker-compose up`.
**Spoken:** *"Engram runs on Go and Redis. It's event-sourced, meaning every thought is an immutable event in a log. It starts in milliseconds."*

---

## Part 3: The "Deep Thought" Crash test (Visual)
**Step 1:** Start a Python script that asks the agent a multi-part question (e.g., "Tell me a long story about a robot named Engram, but pause every few sentences for me to say 'Continue'.").
**Step 2:** While the agent is responding, **KILL THE DOCKER CONTAINER** (Ctrl+C or `docker stop`).
**Step 3:** The Python script will hang/error. This is the "Oh Sh*t" moment.
**Step 4:** Restart the engine: `docker-compose up`.

**Spoken:** *"Now, watch this. I'm killing the engine right in the middle of the agent's response. In any other framework, this agent is dead. But in Engram, we just restart the container. No data loss, no state loss."*

---

## Part 4: The Recovery (Visual)
**Step 1:** Run the Python script again or resume the stream.
**Step 2:** The agent will **resume exactly where it left off**, finishing the sentence it was in the middle of.

**Spoken:** *"Notice how the agent didn't ask 'Where was I?'. It replayed its historical 'soul' from the Redis log and simply continued the story. This is zero-overhead, nanosecond-precision recovery."*

---

## Part 5: Secure WASM Sandboxing (Visual)
**Step 1:** Create an agent and ask it to run a tool (e.g., "Check the system clock").
**Step 2:** Show the engine logs showing: `Sandboxed execution: clock.wasm`.
**Step 3:** Point at the code where the WASM tool resides.

**Spoken:** *"And it's not just durable; it's secure. When this agent needs to execute code, it doesn't run on my os. It spawns a strictly isolated WASM sandbox using wazero. Full CPU and memory isolation with zero lateral access."*

---

## Part 6: Closing (Spoken)
*"Engram is the operating system for the next generation of autonomous colleagues. Durable, secure, and built for the token-per-second era. Check us out on GitHub!"*

---

## 💡 Pro-Tips for the Video:

### Visual Strategy: Show, Don't Write
For a professional demo, **don't write the code live**. It slows down the pace and risks typos.
- **The Staging**: Have the Python script (`agent_demo.py`) already open in your editor.
- **The Highlight**: Highlight the 3-4 lines that matter (the `stub.Chat` stream) and say: *"As you can see, the SDK is just standard gRPC, making it incredibly lightweight."*
- **The Speed**: Use `time.time()` markers in your terminal to show that recovery takes practically zero overhead.

### Key Visuals to Include:
1. **The Logs**: Keep a terminal with Docker logs visible. People love seeing the "Event Log Replay" messages.
2. **The Terminal Split**: Split your screen: Top is the Agent output, Bottom is the Engine logs.
3. **The Reveal**: When the agent recovers, use a high-contrast terminal theme so the "Continued" text is obvious.

### The "Hand-Off" (Advanced)
If you have two workers, show one worker picking up the agent from another worker that died. This proves the shared state via Redis.
