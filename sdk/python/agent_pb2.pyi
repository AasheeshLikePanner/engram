import datetime

from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf import empty_pb2 as _empty_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SetAgentStatusRequest(_message.Message):
    __slots__ = ("agent_id", "status")
    AGENT_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    agent_id: str
    status: str
    def __init__(self, agent_id: _Optional[str] = ..., status: _Optional[str] = ...) -> None: ...

class SetAgentStatusResponse(_message.Message):
    __slots__ = ("status",)
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: str
    def __init__(self, status: _Optional[str] = ...) -> None: ...

class CreateAgentRequest(_message.Message):
    __slots__ = ("goal", "model", "metadata", "tools", "system_prompt", "config")
    class MetadataEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    GOAL_FIELD_NUMBER: _ClassVar[int]
    MODEL_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    TOOLS_FIELD_NUMBER: _ClassVar[int]
    SYSTEM_PROMPT_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    goal: str
    model: str
    metadata: _containers.ScalarMap[str, str]
    tools: _containers.RepeatedCompositeFieldContainer[Tool]
    system_prompt: str
    config: AgentConfig
    def __init__(self, goal: _Optional[str] = ..., model: _Optional[str] = ..., metadata: _Optional[_Mapping[str, str]] = ..., tools: _Optional[_Iterable[_Union[Tool, _Mapping]]] = ..., system_prompt: _Optional[str] = ..., config: _Optional[_Union[AgentConfig, _Mapping]] = ...) -> None: ...

class CreateAgentResponse(_message.Message):
    __slots__ = ("agent_id", "agent")
    AGENT_ID_FIELD_NUMBER: _ClassVar[int]
    AGENT_FIELD_NUMBER: _ClassVar[int]
    agent_id: str
    agent: Agent
    def __init__(self, agent_id: _Optional[str] = ..., agent: _Optional[_Union[Agent, _Mapping]] = ...) -> None: ...

class AgentConfig(_message.Message):
    __slots__ = ("max_steps", "max_duration", "temperature", "max_context_tokens", "enable_thinking", "enable_observation")
    MAX_STEPS_FIELD_NUMBER: _ClassVar[int]
    MAX_DURATION_FIELD_NUMBER: _ClassVar[int]
    TEMPERATURE_FIELD_NUMBER: _ClassVar[int]
    MAX_CONTEXT_TOKENS_FIELD_NUMBER: _ClassVar[int]
    ENABLE_THINKING_FIELD_NUMBER: _ClassVar[int]
    ENABLE_OBSERVATION_FIELD_NUMBER: _ClassVar[int]
    max_steps: int
    max_duration: _duration_pb2.Duration
    temperature: float
    max_context_tokens: int
    enable_thinking: bool
    enable_observation: bool
    def __init__(self, max_steps: _Optional[int] = ..., max_duration: _Optional[_Union[datetime.timedelta, _duration_pb2.Duration, _Mapping]] = ..., temperature: _Optional[float] = ..., max_context_tokens: _Optional[int] = ..., enable_thinking: bool = ..., enable_observation: bool = ...) -> None: ...

class Agent(_message.Message):
    __slots__ = ("agent_id", "goal", "status", "created_at", "updated_at", "metadata", "stats")
    class MetadataEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    AGENT_ID_FIELD_NUMBER: _ClassVar[int]
    GOAL_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    STATS_FIELD_NUMBER: _ClassVar[int]
    agent_id: str
    goal: str
    status: str
    created_at: _timestamp_pb2.Timestamp
    updated_at: _timestamp_pb2.Timestamp
    metadata: _containers.ScalarMap[str, str]
    stats: AgentStats
    def __init__(self, agent_id: _Optional[str] = ..., goal: _Optional[str] = ..., status: _Optional[str] = ..., created_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., updated_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., metadata: _Optional[_Mapping[str, str]] = ..., stats: _Optional[_Union[AgentStats, _Mapping]] = ...) -> None: ...

class AgentStats(_message.Message):
    __slots__ = ("total_steps", "tool_calls", "total_duration", "tokens_used")
    TOTAL_STEPS_FIELD_NUMBER: _ClassVar[int]
    TOOL_CALLS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_DURATION_FIELD_NUMBER: _ClassVar[int]
    TOKENS_USED_FIELD_NUMBER: _ClassVar[int]
    total_steps: int
    tool_calls: int
    total_duration: _duration_pb2.Duration
    tokens_used: int
    def __init__(self, total_steps: _Optional[int] = ..., tool_calls: _Optional[int] = ..., total_duration: _Optional[_Union[datetime.timedelta, _duration_pb2.Duration, _Mapping]] = ..., tokens_used: _Optional[int] = ...) -> None: ...

class ClientMessage(_message.Message):
    __slots__ = ("agent_id", "client_timestamp", "user_input", "pause", "resume", "edit", "cancel")
    AGENT_ID_FIELD_NUMBER: _ClassVar[int]
    CLIENT_TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    USER_INPUT_FIELD_NUMBER: _ClassVar[int]
    PAUSE_FIELD_NUMBER: _ClassVar[int]
    RESUME_FIELD_NUMBER: _ClassVar[int]
    EDIT_FIELD_NUMBER: _ClassVar[int]
    CANCEL_FIELD_NUMBER: _ClassVar[int]
    agent_id: str
    client_timestamp: _timestamp_pb2.Timestamp
    user_input: UserInput
    pause: Pause
    resume: Resume
    edit: Edit
    cancel: Cancel
    def __init__(self, agent_id: _Optional[str] = ..., client_timestamp: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., user_input: _Optional[_Union[UserInput, _Mapping]] = ..., pause: _Optional[_Union[Pause, _Mapping]] = ..., resume: _Optional[_Union[Resume, _Mapping]] = ..., edit: _Optional[_Union[Edit, _Mapping]] = ..., cancel: _Optional[_Union[Cancel, _Mapping]] = ...) -> None: ...

class UserInput(_message.Message):
    __slots__ = ("content", "attachments")
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    ATTACHMENTS_FIELD_NUMBER: _ClassVar[int]
    content: str
    attachments: _containers.RepeatedCompositeFieldContainer[Attachment]
    def __init__(self, content: _Optional[str] = ..., attachments: _Optional[_Iterable[_Union[Attachment, _Mapping]]] = ...) -> None: ...

class Attachment(_message.Message):
    __slots__ = ("name", "mime_type", "data")
    NAME_FIELD_NUMBER: _ClassVar[int]
    MIME_TYPE_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    name: str
    mime_type: str
    data: bytes
    def __init__(self, name: _Optional[str] = ..., mime_type: _Optional[str] = ..., data: _Optional[bytes] = ...) -> None: ...

class Pause(_message.Message):
    __slots__ = ("reason",)
    REASON_FIELD_NUMBER: _ClassVar[int]
    reason: str
    def __init__(self, reason: _Optional[str] = ...) -> None: ...

class Resume(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class Cancel(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class Retry(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class Edit(_message.Message):
    __slots__ = ("patch", "description")
    PATCH_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    patch: _struct_pb2.Struct
    description: str
    def __init__(self, patch: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., description: _Optional[str] = ...) -> None: ...

class ServerMessage(_message.Message):
    __slots__ = ("agent_id", "sequence_id", "timestamp", "thought", "observation", "tool_call", "tool_result", "text", "final_answer", "error", "status", "heartbeat")
    AGENT_ID_FIELD_NUMBER: _ClassVar[int]
    SEQUENCE_ID_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    THOUGHT_FIELD_NUMBER: _ClassVar[int]
    OBSERVATION_FIELD_NUMBER: _ClassVar[int]
    TOOL_CALL_FIELD_NUMBER: _ClassVar[int]
    TOOL_RESULT_FIELD_NUMBER: _ClassVar[int]
    TEXT_FIELD_NUMBER: _ClassVar[int]
    FINAL_ANSWER_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    HEARTBEAT_FIELD_NUMBER: _ClassVar[int]
    agent_id: str
    sequence_id: int
    timestamp: _timestamp_pb2.Timestamp
    thought: Thought
    observation: Observation
    tool_call: ToolCall
    tool_result: ToolResult
    text: TextOutput
    final_answer: FinalAnswer
    error: Error
    status: StatusUpdate
    heartbeat: Heartbeat
    def __init__(self, agent_id: _Optional[str] = ..., sequence_id: _Optional[int] = ..., timestamp: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., thought: _Optional[_Union[Thought, _Mapping]] = ..., observation: _Optional[_Union[Observation, _Mapping]] = ..., tool_call: _Optional[_Union[ToolCall, _Mapping]] = ..., tool_result: _Optional[_Union[ToolResult, _Mapping]] = ..., text: _Optional[_Union[TextOutput, _Mapping]] = ..., final_answer: _Optional[_Union[FinalAnswer, _Mapping]] = ..., error: _Optional[_Union[Error, _Mapping]] = ..., status: _Optional[_Union[StatusUpdate, _Mapping]] = ..., heartbeat: _Optional[_Union[Heartbeat, _Mapping]] = ...) -> None: ...

class Thought(_message.Message):
    __slots__ = ("content",)
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    content: str
    def __init__(self, content: _Optional[str] = ...) -> None: ...

class Observation(_message.Message):
    __slots__ = ("content",)
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    content: str
    def __init__(self, content: _Optional[str] = ...) -> None: ...

class ToolCall(_message.Message):
    __slots__ = ("tool_name", "input_json", "call_id")
    TOOL_NAME_FIELD_NUMBER: _ClassVar[int]
    INPUT_JSON_FIELD_NUMBER: _ClassVar[int]
    CALL_ID_FIELD_NUMBER: _ClassVar[int]
    tool_name: str
    input_json: str
    call_id: str
    def __init__(self, tool_name: _Optional[str] = ..., input_json: _Optional[str] = ..., call_id: _Optional[str] = ...) -> None: ...

class ToolResult(_message.Message):
    __slots__ = ("call_id", "output", "is_error", "error_message")
    CALL_ID_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    IS_ERROR_FIELD_NUMBER: _ClassVar[int]
    ERROR_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    call_id: str
    output: str
    is_error: bool
    error_message: str
    def __init__(self, call_id: _Optional[str] = ..., output: _Optional[str] = ..., is_error: bool = ..., error_message: _Optional[str] = ...) -> None: ...

class TextOutput(_message.Message):
    __slots__ = ("content",)
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    content: str
    def __init__(self, content: _Optional[str] = ...) -> None: ...

class FinalAnswer(_message.Message):
    __slots__ = ("content", "structured")
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    STRUCTURED_FIELD_NUMBER: _ClassVar[int]
    content: str
    structured: _struct_pb2.Struct
    def __init__(self, content: _Optional[str] = ..., structured: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class Error(_message.Message):
    __slots__ = ("code", "message")
    CODE_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    code: str
    message: str
    def __init__(self, code: _Optional[str] = ..., message: _Optional[str] = ...) -> None: ...

class StatusUpdate(_message.Message):
    __slots__ = ("status", "message")
    STATUS_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    status: str
    message: str
    def __init__(self, status: _Optional[str] = ..., message: _Optional[str] = ...) -> None: ...

class Heartbeat(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetAgentRequest(_message.Message):
    __slots__ = ("agent_id",)
    AGENT_ID_FIELD_NUMBER: _ClassVar[int]
    agent_id: str
    def __init__(self, agent_id: _Optional[str] = ...) -> None: ...

class ListAgentsRequest(_message.Message):
    __slots__ = ("status_filter", "page_size", "page_token")
    STATUS_FILTER_FIELD_NUMBER: _ClassVar[int]
    PAGE_SIZE_FIELD_NUMBER: _ClassVar[int]
    PAGE_TOKEN_FIELD_NUMBER: _ClassVar[int]
    status_filter: str
    page_size: int
    page_token: str
    def __init__(self, status_filter: _Optional[str] = ..., page_size: _Optional[int] = ..., page_token: _Optional[str] = ...) -> None: ...

class ListAgentsResponse(_message.Message):
    __slots__ = ("agents", "next_page_token")
    AGENTS_FIELD_NUMBER: _ClassVar[int]
    NEXT_PAGE_TOKEN_FIELD_NUMBER: _ClassVar[int]
    agents: _containers.RepeatedCompositeFieldContainer[Agent]
    next_page_token: str
    def __init__(self, agents: _Optional[_Iterable[_Union[Agent, _Mapping]]] = ..., next_page_token: _Optional[str] = ...) -> None: ...

class DeleteAgentRequest(_message.Message):
    __slots__ = ("agent_id", "force")
    AGENT_ID_FIELD_NUMBER: _ClassVar[int]
    FORCE_FIELD_NUMBER: _ClassVar[int]
    agent_id: str
    force: bool
    def __init__(self, agent_id: _Optional[str] = ..., force: bool = ...) -> None: ...

class ControlAgentRequest(_message.Message):
    __slots__ = ("agent_id", "pause", "resume", "cancel", "retry")
    AGENT_ID_FIELD_NUMBER: _ClassVar[int]
    PAUSE_FIELD_NUMBER: _ClassVar[int]
    RESUME_FIELD_NUMBER: _ClassVar[int]
    CANCEL_FIELD_NUMBER: _ClassVar[int]
    RETRY_FIELD_NUMBER: _ClassVar[int]
    agent_id: str
    pause: Pause
    resume: Resume
    cancel: Cancel
    retry: Retry
    def __init__(self, agent_id: _Optional[str] = ..., pause: _Optional[_Union[Pause, _Mapping]] = ..., resume: _Optional[_Union[Resume, _Mapping]] = ..., cancel: _Optional[_Union[Cancel, _Mapping]] = ..., retry: _Optional[_Union[Retry, _Mapping]] = ...) -> None: ...

class ControlAgentResponse(_message.Message):
    __slots__ = ("status",)
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: str
    def __init__(self, status: _Optional[str] = ...) -> None: ...

class ReplayRequest(_message.Message):
    __slots__ = ("agent_id", "from_sequence", "include_history")
    AGENT_ID_FIELD_NUMBER: _ClassVar[int]
    FROM_SEQUENCE_FIELD_NUMBER: _ClassVar[int]
    INCLUDE_HISTORY_FIELD_NUMBER: _ClassVar[int]
    agent_id: str
    from_sequence: int
    include_history: bool
    def __init__(self, agent_id: _Optional[str] = ..., from_sequence: _Optional[int] = ..., include_history: bool = ...) -> None: ...

class Tool(_message.Message):
    __slots__ = ("name", "description", "parameters_schema", "requires_confirmation")
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    PARAMETERS_SCHEMA_FIELD_NUMBER: _ClassVar[int]
    REQUIRES_CONFIRMATION_FIELD_NUMBER: _ClassVar[int]
    name: str
    description: str
    parameters_schema: _struct_pb2.Struct
    requires_confirmation: bool
    def __init__(self, name: _Optional[str] = ..., description: _Optional[str] = ..., parameters_schema: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., requires_confirmation: bool = ...) -> None: ...
