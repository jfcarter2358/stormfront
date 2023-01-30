package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	CurrentCluster string          `json:"current_cluster" yaml:"current_cluster"`
	Clusters       []ClusterConfig `json:"clusters" yaml:"clusters"`
}

type ClusterConfig struct {
	Token            string   `json:"token" yaml:"token"`                         // Token used for authentication
	Name             string   `json:"name" yaml:"name"`                           // Name of the cluster to interact with
	CurrentNamespace string   `json:"current_namespace" yaml:"current_namespace"` // Current namespace in cluster
	Namespaces       []string `json:"namespaces" yaml:"namespaces"`               // All available namespaces within the cluster
	Host             string   `json:"host" yaml:"host"`                           // Leader host
	Port             string   `json:"port" yaml:"port"`                           // Leader port
}

func ReadConfig() (Config, error) {
	configPath := getEnvDefault("STORMFRONTCONFIG", "~/.stormfrontconfig")
	_, err := os.Stat(configPath)
	if err == nil {
		var data Config
		file, _ := ioutil.ReadFile("conf.yaml")
		yaml.Unmarshal(file, &data)
		return data, nil
	}
	return Config{CurrentCluster: "", Clusters: []ClusterConfig{}}, nil
}

func WriteConfig(config Config) error {
	configPath := getEnvDefault("STORMFRONTCONFIG", "~/.stormfrontconfig")

	data, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(configPath, data, 0)
	if err != nil {
		return err
	}

	return nil
}

func ChangeCluster(newCluster string) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}

	config.CurrentCluster = newCluster

	err = WriteConfig(config)

	return err
}

func AddCluster(clusterConfig ClusterConfig) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}

	config.Clusters = append(config.Clusters, clusterConfig)

	err = WriteConfig(config)

	return err
}

func RemoveCluster(clusterName string) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}

	for idx, cluster := range config.Clusters {
		if cluster.Name == clusterName {
			config.Clusters = append(config.Clusters[:idx], config.Clusters[idx+1:]...)
			break
		}
	}

	err = WriteConfig(config)

	return err
}

func GetCluster() (string, error) {
	config, err := ReadConfig()
	if err != nil {
		return "", err
	}

	return config.CurrentCluster, nil
}

func GetClusters() ([]ClusterConfig, error) {
	config, err := ReadConfig()
	if err != nil {
		return []ClusterConfig{}, err
	}

	return config.Clusters, nil
}

func RenameCluster(oldName, newName string) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}

	if oldName == config.CurrentCluster {
		config.CurrentCluster = newName
	}

	for idx, cluster := range config.Clusters {
		if cluster.Name == oldName {
			config.Clusters[idx].Name = newName
		}
	}

	err = WriteConfig(config)

	return err
}

func ChangeNamespace(name string) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}

	for idx, cluster := range config.Clusters {
		if cluster.Name == config.CurrentCluster {
			config.Clusters[idx].CurrentNamespace = name
			break
		}
	}

	err = WriteConfig(config)

	return err
}

func GetNamespace() (string, error) {
	config, err := ReadConfig()
	if err != nil {
		return "", err
	}

	for _, cluster := range config.Clusters {
		if cluster.Name == config.CurrentCluster {
			return cluster.CurrentNamespace, nil
		}
	}

	return "", fmt.Errorf("could not find entry for cluster %s", config.CurrentCluster)
}

func GetNamespaces() ([]string, error) {
	config, err := ReadConfig()
	if err != nil {
		return []string{}, err
	}

	for _, cluster := range config.Clusters {
		if cluster.Name == config.CurrentCluster {
			return cluster.Namespaces, nil
		}
	}

	return []string{}, fmt.Errorf("could not find entry for cluster %s", config.CurrentCluster)
}

func AddNamespace(name string) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}

	for idx, clusterConfig := range config.Clusters {
		if config.CurrentCluster == clusterConfig.Name {
			config.Clusters[idx].Namespaces = append(config.Clusters[idx].Namespaces, name)
		}
	}

	err = WriteConfig(config)

	return err
}

func RemoveNamespace(name string) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}

	for idx, cluster := range config.Clusters {
		if cluster.Name == config.CurrentCluster {
			for jdx, namespace := range cluster.Namespaces {
				if namespace == name {
					config.Clusters[idx].Namespaces = append(config.Clusters[idx].Namespaces[:jdx], config.Clusters[idx].Namespaces[jdx+1:]...)
					break
				}
			}
			break
		}
	}

	err = WriteConfig(config)

	return err
}

func GetAPIToken() (string, error) {
	config, err := ReadConfig()
	if err != nil {
		return "", err
	}
	for _, cluster := range config.Clusters {
		if cluster.Name == config.CurrentCluster {
			return cluster.Token, nil
		}
	}

	return "", fmt.Errorf("could not find entry for cluster %s", config.CurrentCluster)
}

func GetHost() (string, error) {
	config, err := ReadConfig()
	if err != nil {
		return "", err
	}

	for _, cluster := range config.Clusters {
		if cluster.Name == config.CurrentCluster {
			return cluster.Host, nil
		}
	}

	return "", fmt.Errorf("could not find entry for cluster %s", config.CurrentCluster)
}

func GetPort() (string, error) {
	config, err := ReadConfig()
	if err != nil {
		return "", err
	}

	for _, cluster := range config.Clusters {
		if cluster.Name == config.CurrentCluster {
			return cluster.Port, nil
		}
	}

	return "", fmt.Errorf("could not find entry for cluster %s", config.CurrentCluster)
}

func getEnvDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}

	return value
}
