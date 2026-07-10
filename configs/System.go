package configs

type SystemConfig struct {
	Url          string `json:"url"`
	ApiKey       string `json:"api_key"`
	SystemPrompt string `json:"system_prompt"`
	CurDir       string `json:"cur_dir"`
}

type ModelConfig struct {
	Model     string `json:"model"`
	MaxTokens int64  `json:"maxTokens"`
}

var SysCfg SystemConfig
var ModelCfg ModelConfig
