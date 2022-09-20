package services

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type FileManagerTestSuite struct {
	suite.Suite
	logger       *zap.Logger
	filesManager FileManager
	folder       string
}

func TestFileManagerSuite(t *testing.T) {
	s := new(FileManagerTestSuite)
	suite.Run(t, s)
}

func (s *FileManagerTestSuite) SetupTest() {
	s.logger = zap.NewExample()
	s.folder = "test_folder"
	service := NewFileManager(s.folder, s.logger)
	s.filesManager = service
}

func (s *FileManagerTestSuite) TestFileManager_GenerateFileName_OK() {
	t := s.T()
	isLettersAndDigits := regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString
	nameReturn, err := s.filesManager.GenerateFileName()
	assert.NoError(t, err)
	assert.True(t, isLettersAndDigits(nameReturn))
}

func (s *FileManagerTestSuite) TestFileManager_SaveFile_OK() {
	t := s.T()
	fileName := "testfile.txt"

	err := s.filesManager.SaveDataToFile([]byte{1, 1, 1, 1}, fileName)
	assert.NoError(t, err)

	_, err = os.Stat(filepath.Join(s.folder, fileName))
	assert.NoError(t, err)

	s.cleanResources()
}

func (s *FileManagerTestSuite) TestFileManager_ReadFile_NotExists() {
	t := s.T()
	fileName := "testfile.txt"

	returnedBytes, err := s.filesManager.ReadFile(fileName)
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
	assert.Equal(t, []byte(nil), returnedBytes)
}

func (s *FileManagerTestSuite) TestFileManager_ReadFile_Empty() {
	t := s.T()
	fileName := "testfile.txt"

	f, err := s.createFileForTest(fileName)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()

	returnedBytes, err := s.filesManager.ReadFile(fileName)
	assert.NoError(t, err)
	assert.Equal(t, []byte{}, returnedBytes)
	assert.Equal(t, 0, len(returnedBytes))

	s.cleanResources()
}

func (s *FileManagerTestSuite) TestFileManager_ReadFile_OK() {
	t := s.T()
	fileName := "testfile.txt"
	dataStr := "Hello, test"

	f, err := s.createFileForTest(fileName)
	if err != nil {
		log.Fatal(err)
	}

	n, err := f.Write([]byte(dataStr))
	if err != nil {
		log.Fatal(err)
	}
	f.Close()

	returnedBytes, err := s.filesManager.ReadFile(fileName)
	assert.NoError(t, err)
	assert.Equal(t, []byte(dataStr), returnedBytes)
	assert.Equal(t, n, len(returnedBytes))

	s.cleanResources()
}

func (s *FileManagerTestSuite) TestFileManager_DeleteFile_NotExists() {
	t := s.T()
	fileName := "testfile.txt"

	err := s.filesManager.DeleteFile(fileName)
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func (s *FileManagerTestSuite) TestFileManager_DeleteFile_OK() {
	t := s.T()
	fileName := "testfile.txt"

	f, err := s.createFileForTest(fileName)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()

	err = s.filesManager.DeleteFile(fileName)
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(s.folder, fileName))
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))

	s.cleanResources()
}

func (s *FileManagerTestSuite) createFileForTest(fileName string) (*os.File, error) {
	if _, err := os.Stat(s.folder); os.IsNotExist(err) {
		err = os.Mkdir(s.folder, 0766)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	f, err := os.Create(filepath.Join(s.folder, fileName))
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (s *FileManagerTestSuite) cleanResources() error {
	return os.RemoveAll(s.folder)
}
