package services

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	BuildFileURL(server, path, name string) (string, error)
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

func (f *fileManager) BuildFileURL(server, pathURL, name string) (string, error) {
	u, err := url.Parse(server)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, pathURL, name)
	return u.String(), nil
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
