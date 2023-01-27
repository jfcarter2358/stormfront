package client

type StormfrontNode struct {
	ID     string               `json:"id" yaml:"id"`
	Host   string               `json:"host" yaml:"host"`
	Port   int                  `json:"port" yaml:"port"`
	System StormfrontSystemInfo `json:"system" yaml:"system"`
	Health string               `json:"health" yaml:"health"`
	Type   string               `json:"type" yaml:"type"`
}
