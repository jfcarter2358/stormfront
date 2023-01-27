package client

type StormfrontLeader struct {
	ID         string   `json:"id" yaml:"id"`
	Succession []string `json:"succession" yaml:"succession"`
	Healthy    []string `json:"healthy" yaml:"healthy"`
	Unhealthy  []string `json:"unhealthy" yaml:"unhealthy"`
	Unknown    []string `json:"unknown" yaml:"unknown"`
}
