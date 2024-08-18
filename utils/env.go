package utils

import (
	"errors"
	"os"
)

func GetEnv(key string) (string, error) {
	data, found := os.LookupEnv(key)
	if !found {
		return "", errors.New("key not found")
	}
	return data, nil
}
