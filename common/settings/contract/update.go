package contract

type BatchUpdateConfig struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	ConfigType string `json:"configType"`
}
