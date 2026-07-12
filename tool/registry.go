package tool

import (
	"go-agent/services"
)

func RegisterTools(req *services.ChatRequest) {
	registerToolBash(req)
	registerToolFileSystem(req)
}
