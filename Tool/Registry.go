package Tool

import "go-agent/Services"

func RegisterTools(req *Services.ChatRequest) {
	registerToolBash(req)
	registerToolFileSystem(req)
}
