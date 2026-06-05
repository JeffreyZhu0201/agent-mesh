package config

type Config struct {
	Name  string `yaml:"name"`
	Port  int    `yaml:"port"`
	MySQL struct {
		Host            string `yaml:"host"`
		Port            int    `yaml:"port"`
		User            string `yaml:"user"`
		Password        string `yaml:"password"`
		Database        string `yaml:"database"`
		MaxOpenConns    int    `yaml:"maxOpenConns"`
		MaxIdleConns    int    `yaml:"maxIdleConns"`
		ConnMaxLifetime int    `yaml:"connMaxLifetime"`
	} `yaml:"mysql"`
	JWT struct {
		Secret  string `yaml:"secret"`
		SignKey string `yaml:"signKey"`
		Expiry  int    `yaml:"expiry"`
	} `yaml:"jwt"`
	LLM struct {
		Provider string `yaml:"provider"` // "openai", "claude", "ollama"
		OpenAI   struct {
			APIKey   string `yaml:"apiKey"`
			Endpoint string `yaml:"endpoint"`
			Model    string `yaml:"model"`
		} `yaml:"openai"`
		Claude struct {
			APIKey   string `yaml:"apiKey"`
			Endpoint string `yaml:"endpoint"`
			Model    string `yaml:"model"`
		} `yaml:"claude"`
		Ollama struct {
			Endpoint string `yaml:"endpoint"`
			Model    string `yaml:"model"`
		} `yaml:"ollama"`
		DefaultModel    string  `yaml:"defaultModel"`
		Temperature     float32 `yaml:"temperature"`
		MaxTokens       int     `yaml:"maxTokens"`
	} `yaml:"llm"`
}
