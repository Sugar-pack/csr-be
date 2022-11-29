package utils

import (
	"strings"

	"github.com/google/uuid"
)

func GenerateFileName() (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	name := strings.Replace(id.String(), "-", "", -1)
	return name, nil
}
