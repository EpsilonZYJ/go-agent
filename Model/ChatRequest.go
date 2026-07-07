package Model

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"message"`
	Tools    []Tool    `json:"tools"`
}
