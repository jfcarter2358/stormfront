package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

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
	usr, err := user.Current()
	if err != nil {
		return Config{}, err
	}
	confPath := getEnvDefault("STORMFRONTCONFIG", fmt.Sprintf("%s/.stormfrontconfig", usr.HomeDir))
	_, err = os.Stat(confPath)
	if err == nil {
		var data Config
		file, err := ioutil.ReadFile(confPath)
		if err != nil {
			return Config{}, err
		}
		yaml.Unmarshal(file, &data)
		return data, nil
	}
	return Config{CurrentCluster: "", Clusters: []ClusterConfig{}}, nil
}

func WriteConfig(conf Config) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	confPath := getEnvDefault("STORMFRONTCONFIG", fmt.Sprintf("%s/.stormfrontconfig", usr.HomeDir))

	data, err := yaml.Marshal(&conf)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(confPath, data, 0740)
	if err != nil {
		return err
	}

	return nil
}

func ChangeCluster(newCluster string) error {
	conf, err := ReadConfig()
	if err != nil {
		return err
	}

	conf.CurrentCluster = newCluster

	err = WriteConfig(conf)

	return err
}

func AddCluster(clusterConfig ClusterConfig) error {
	conf, err := ReadConfig()
	if err != nil {
		return err
	}

	conf.Clusters = append(conf.Clusters, clusterConfig)

	err = WriteConfig(conf)

	return err
}

func RemoveCluster(clusterName string) error {
	conf, err := ReadConfig()
	if err != nil {
		return err
	}

	for idx, cluster := range conf.Clusters {
		if cluster.Name == clusterName {
			conf.Clusters = append(conf.Clusters[:idx], conf.Clusters[idx+1:]...)
			break
		}
	}

	err = WriteConfig(conf)

	return err
}

func GetCluster() (string, error) {
	conf, err := ReadConfig()
	if err != nil {
		return "", err
	}

	return conf.CurrentCluster, nil
}

func GetClusters() ([]ClusterConfig, error) {
	conf, err := ReadConfig()
	if err != nil {
		return []ClusterConfig{}, err
	}

	return conf.Clusters, nil
}

func RenameCluster(oldName, newName string) error {
	conf, err := ReadConfig()
	if err != nil {
		return err
	}

	if oldName == conf.CurrentCluster {
		conf.CurrentCluster = newName
	}

	for idx, cluster := range conf.Clusters {
		if cluster.Name == oldName {
			conf.Clusters[idx].Name = newName
		}
	}

	err = WriteConfig(conf)

	return err
}

func ChangeNamespace(name string) error {
	conf, err := ReadConfig()
	if err != nil {
		return err
	}

	for idx, cluster := range conf.Clusters {
		if cluster.Name == conf.CurrentCluster {
			conf.Clusters[idx].CurrentNamespace = name
			break
		}
	}

	err = WriteConfig(conf)

	return err
}

func GetNamespace() (string, error) {
	conf, err := ReadConfig()
	if err != nil {
		return "", err
	}

	for _, cluster := range conf.Clusters {
		if cluster.Name == conf.CurrentCluster {
			return cluster.CurrentNamespace, nil
		}
	}

	return "", fmt.Errorf("could not find entry for cluster %s", conf.CurrentCluster)
}

func GetNamespaces() ([]string, error) {
	conf, err := ReadConfig()
	if err != nil {
		return []string{}, err
	}

	for _, cluster := range conf.Clusters {
		if cluster.Name == conf.CurrentCluster {
			return cluster.Namespaces, nil
		}
	}

	return []string{}, fmt.Errorf("could not find entry for cluster %s", conf.CurrentCluster)
}

func AddNamespace(name string) error {
	conf, err := ReadConfig()
	if err != nil {
		return err
	}

	for idx, clusterConfig := range conf.Clusters {
		if conf.CurrentCluster == clusterConfig.Name {
			conf.Clusters[idx].Namespaces = append(conf.Clusters[idx].Namespaces, name)
		}
	}

	err = WriteConfig(conf)

	return err
}

func RemoveNamespace(name string) error {
	conf, err := ReadConfig()
	if err != nil {
		return err
	}

	for idx, cluster := range conf.Clusters {
		if cluster.Name == conf.CurrentCluster {
			for jdx, namespace := range cluster.Namespaces {
				if namespace == name {
					conf.Clusters[idx].Namespaces = append(conf.Clusters[idx].Namespaces[:jdx], conf.Clusters[idx].Namespaces[jdx+1:]...)
					break
				}
			}
			break
		}
	}

	err = WriteConfig(conf)

	return err
}

func GetAPIToken() (string, error) {
	conf, err := ReadConfig()
	if err != nil {
		return "", err
	}
	for _, cluster := range conf.Clusters {
		if cluster.Name == conf.CurrentCluster {
			return cluster.Token, nil
		}
	}

	return "", fmt.Errorf("could not find entry for cluster %s", conf.CurrentCluster)
}

func GetHost() (string, error) {
	conf, err := ReadConfig()
	if err != nil {
		return "", err
	}

	for _, cluster := range conf.Clusters {
		if cluster.Name == conf.CurrentCluster {
			return cluster.Host, nil
		}
	}

	return "", fmt.Errorf("could not find entry for cluster %s", conf.CurrentCluster)
}

func GetPort() (string, error) {
	conf, err := ReadConfig()
	if err != nil {
		return "", err
	}

	for _, cluster := range conf.Clusters {
		if cluster.Name == conf.CurrentCluster {
			return cluster.Port, nil
		}
	}

	return "", fmt.Errorf("could not find entry for cluster %s", conf.CurrentCluster)
}

func getEnvDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}

	return value
}
