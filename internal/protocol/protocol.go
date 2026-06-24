package protocol

import "encoding/json"

type MessageType string

const (
	TypeHandshake   MessageType = "handshake"
	TypeHandshakeOK MessageType = "handshake_ok"
	TypeEnroll      MessageType = "enroll"
	TypeEnrolled    MessageType = "enrolled"
	TypeStdout      MessageType = "stdout"
	TypeStderr      MessageType = "stderr"
	TypeDone        MessageType = "done"
	TypeError       MessageType = "error"
	TypeTask        MessageType = "task"
)

type Message struct {
	Type    MessageType `json:"type"`
	TaskID  string      `json:"task_id,omitempty"`
	Content string      `json:"content,omitempty"`

	// Handshake (daemon → gateway)
	DaemonID  string   `json:"daemon_id,omitempty"`
	Token     string   `json:"token,omitempty"`
	Harnesses []string `json:"harnesses,omitempty"`

	// Enroll (daemon → gateway)
	WorkspaceKey string `json:"workspace_key,omitempty"`
	DaemonName   string `json:"daemon_name,omitempty"`

	// Enrolled (gateway → daemon)
	EnrolledID    string `json:"enrolled_id,omitempty"`
	EnrolledToken string `json:"enrolled_token,omitempty"`

	// Task (gateway → daemon)
	Harness   string   `json:"harness,omitempty"`
	Prompt    string   `json:"prompt,omitempty"`
	Model     string   `json:"model,omitempty"`
	ExtraArgs []string `json:"extra_args,omitempty"`
}

func (m Message) Encode() ([]byte, error) { return json.Marshal(m) }
func DecodeMessage(data []byte) (Message, error) {
	var m Message
	return m, json.Unmarshal(data, &m)
}
