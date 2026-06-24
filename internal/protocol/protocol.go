package protocol

import "encoding/json"

type MessageType string

const (
	TypeHandshake   MessageType = "handshake"
	TypeHandshakeOK MessageType = "handshake_ok"
	TypeStreamEvent MessageType = "stream_event"
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
	DaemonID   string   `json:"daemon_id,omitempty"`
	Token      string   `json:"token,omitempty"`
	DaemonName string   `json:"daemon_name,omitempty"`
	Harnesses  []string            `json:"harnesses,omitempty"`
	HarnessModels map[string][]string `json:"harness_models,omitempty"`

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
