import datetime

from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AgentMetadata(_message.Message):
    __slots__ = ("agent_id", "status", "running_on_worker", "step_started_at", "updated_at", "head_sequence", "current_model")
    AGENT_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    RUNNING_ON_WORKER_FIELD_NUMBER: _ClassVar[int]
    STEP_STARTED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    HEAD_SEQUENCE_FIELD_NUMBER: _ClassVar[int]
    CURRENT_MODEL_FIELD_NUMBER: _ClassVar[int]
    agent_id: str
    status: str
    running_on_worker: str
    step_started_at: _timestamp_pb2.Timestamp
    updated_at: _timestamp_pb2.Timestamp
    head_sequence: int
    current_model: str
    def __init__(self, agent_id: _Optional[str] = ..., status: _Optional[str] = ..., running_on_worker: _Optional[str] = ..., step_started_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., updated_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., head_sequence: _Optional[int] = ..., current_model: _Optional[str] = ...) -> None: ...

class Event(_message.Message):
    __slots__ = ("sequence", "timestamp", "created", "user_message", "thought", "llm_response", "tool_call", "tool_result", "observation", "final_answer", "paused", "resumed", "edited", "error")
    SEQUENCE_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    CREATED_FIELD_NUMBER: _ClassVar[int]
    USER_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    THOUGHT_FIELD_NUMBER: _ClassVar[int]
    LLM_RESPONSE_FIELD_NUMBER: _ClassVar[int]
    TOOL_CALL_FIELD_NUMBER: _ClassVar[int]
    TOOL_RESULT_FIELD_NUMBER: _ClassVar[int]
    OBSERVATION_FIELD_NUMBER: _ClassVar[int]
    FINAL_ANSWER_FIELD_NUMBER: _ClassVar[int]
    PAUSED_FIELD_NUMBER: _ClassVar[int]
    RESUMED_FIELD_NUMBER: _ClassVar[int]
    EDITED_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    sequence: int
    timestamp: _timestamp_pb2.Timestamp
    created: AgentCreated
    user_message: UserMessage
    thought: AgentThought
    llm_response: LLMResponse
    tool_call: ToolCall
    tool_result: ToolResult
    observation: Observation
    final_answer: FinalAnswer
    paused: AgentPaused
    resumed: AgentResumed
    edited: AgentEdited
    error: Error
    def __init__(self, sequence: _Optional[int] = ..., timestamp: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., created: _Optional[_Union[AgentCreated, _Mapping]] = ..., user_message: _Optional[_Union[UserMessage, _Mapping]] = ..., thought: _Optional[_Union[AgentThought, _Mapping]] = ..., llm_response: _Optional[_Union[LLMResponse, _Mapping]] = ..., tool_call: _Optional[_Union[ToolCall, _Mapping]] = ..., tool_result: _Optional[_Union[ToolResult, _Mapping]] = ..., observation: _Optional[_Union[Observation, _Mapping]] = ..., final_answer: _Optional[_Union[FinalAnswer, _Mapping]] = ..., paused: _Optional[_Union[AgentPaused, _Mapping]] = ..., resumed: _Optional[_Union[AgentResumed, _Mapping]] = ..., edited: _Optional[_Union[AgentEdited, _Mapping]] = ..., error: _Optional[_Union[Error, _Mapping]] = ...) -> None: ...

class AgentCreated(_message.Message):
    __slots__ = ("goal", "model")
    GOAL_FIELD_NUMBER: _ClassVar[int]
    MODEL_FIELD_NUMBER: _ClassVar[int]
    goal: str
    model: str
    def __init__(self, goal: _Optional[str] = ..., model: _Optional[str] = ...) -> None: ...

class UserMessage(_message.Message):
    __slots__ = ("content",)
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    content: str
    def __init__(self, content: _Optional[str] = ...) -> None: ...

class AgentThought(_message.Message):
    __slots__ = ("content",)
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    content: str
    def __init__(self, content: _Optional[str] = ...) -> None: ...

class LLMResponse(_message.Message):
    __slots__ = ("content",)
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    content: str
    def __init__(self, content: _Optional[str] = ...) -> None: ...

class ToolCall(_message.Message):
    __slots__ = ("name", "input", "call_id")
    NAME_FIELD_NUMBER: _ClassVar[int]
    INPUT_FIELD_NUMBER: _ClassVar[int]
    CALL_ID_FIELD_NUMBER: _ClassVar[int]
    name: str
    input: str
    call_id: str
    def __init__(self, name: _Optional[str] = ..., input: _Optional[str] = ..., call_id: _Optional[str] = ...) -> None: ...

class ToolResult(_message.Message):
    __slots__ = ("call_id", "output", "is_error")
    CALL_ID_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    IS_ERROR_FIELD_NUMBER: _ClassVar[int]
    call_id: str
    output: str
    is_error: bool
    def __init__(self, call_id: _Optional[str] = ..., output: _Optional[str] = ..., is_error: bool = ...) -> None: ...

class Observation(_message.Message):
    __slots__ = ("content",)
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    content: str
    def __init__(self, content: _Optional[str] = ...) -> None: ...

class FinalAnswer(_message.Message):
    __slots__ = ("content",)
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    content: str
    def __init__(self, content: _Optional[str] = ...) -> None: ...

class AgentPaused(_message.Message):
    __slots__ = ("reason",)
    REASON_FIELD_NUMBER: _ClassVar[int]
    reason: str
    def __init__(self, reason: _Optional[str] = ...) -> None: ...

class AgentResumed(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class AgentEdited(_message.Message):
    __slots__ = ("patch", "description")
    PATCH_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    patch: _struct_pb2.Struct
    description: str
    def __init__(self, patch: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., description: _Optional[str] = ...) -> None: ...

class Error(_message.Message):
    __slots__ = ("message",)
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    message: str
    def __init__(self, message: _Optional[str] = ...) -> None: ...
