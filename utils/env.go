package utils

import (
	"fmt"
	"os"
)

func GetEnv(key string) (string, error) {
	data, found := os.LookupEnv(key)
	if !found {
		return "", fmt.Errorf("failed to get env. %s not found", key)
	}
	return data, nil
}
