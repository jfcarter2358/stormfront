package client

type StormfrontRoute struct {
	ID        string `json:"id" yaml:"id"`
	Alias     string `json:"alias" yaml:"alias"`
	Hostname  string `json:"hostname" yaml:"hostname"`
	Port      int    `json:"port" yaml:"port"`
	Namespace string `json:"namespace" yaml:"namespace"`
}
