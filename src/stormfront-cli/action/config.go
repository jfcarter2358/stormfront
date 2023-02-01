package action

import "stormfront-cli/config"

func GetConnectionDetails() (string, string, error) {
	host, err := config.GetHost()
	if err != nil {
		return "", "", err
	}
	port, err := config.GetPort()
	if err != nil {
		return "", "", err
	}
	return host, port, nil
}
