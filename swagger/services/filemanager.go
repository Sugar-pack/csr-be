package services

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type fileManager struct {
	folder string
	logger *zap.Logger
}

func NewFileManager(folder string, logger *zap.Logger) FileManager {
	return &fileManager{
		folder: folder,
		logger: logger,
	}
}

type FileManager interface {
	GenerateFileName() (string, error)
	SaveDataToFile(data []byte, name string) error
	ReadFile(name string) ([]byte, error)
	DeleteFile(name string) error
}

func (f *fileManager) GenerateFileName() (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	name := strings.Replace(id.String(), "-", "", -1)
	return name, nil
}

func (f *fileManager) SaveDataToFile(data []byte, name string) error {
	if _, err := os.Stat(f.folder); os.IsNotExist(err) {
		err = os.Mkdir(f.folder, 0766)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	err := ioutil.WriteFile(
		filepath.Join(f.folder, name),
		data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (f *fileManager) ReadFile(name string) ([]byte, error) {
	fileBytes, err := ioutil.ReadFile(filepath.Join(f.folder, name))
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}

func (f *fileManager) DeleteFile(name string) error {
	err := os.Remove(filepath.Join(f.folder, name))
	if err != nil {
		return err
	}
	return nil
}
