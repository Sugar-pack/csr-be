package handlers

import (
	"encoding/json"
	"fmt"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/photos"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	photosFolder   string
	photoServerURL string
	photoURLPath   = "api/equipment/photos/"
)

type Photo struct {
	logger *zap.Logger
}

func NewPhoto(folder, serverURL string, logger *zap.Logger) *Photo {
	photosFolder = folder
	photoServerURL = serverURL
	return &Photo{
		logger: logger,
	}
}

func (p Photo) CreateNewPhotoFunc(repository repositories.PhotoRepository) photos.CreateNewPhotoHandlerFunc {
	return func(s photos.CreateNewPhotoParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		photoName, err := generatePhotoName()
		if err != nil {
			p.logger.Error("failed to generate photo name", zap.Error(err))
			return photos.NewCreateNewPhotoDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}

		if err := savePhotoFile(s.File, photoName); err != nil {
			p.logger.Error("failed to save photo to file", zap.Error(err))
			return photos.NewCreateNewPhotoDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}
		photoURL, err := getPhotoURL(photoName)
		if err != nil {
			p.logger.Error("failed to get photo url", zap.Error(err))
			return photos.NewCreateNewPhotoDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}
		newPhoto := models.Photo{
			ID:  photoName,
			URL: &photoURL,
		}
		_, err = repository.CreatePhoto(ctx, newPhoto)
		if err != nil {
			p.logger.Error("failed to save photo to db", zap.Error(err))
			return photos.NewCreateNewPhotoDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}
		return photos.NewCreateNewPhotoCreated().WithPayload(&models.CreateNewPhotoResponse{
			Data: &newPhoto,
		})
	}
}

func (p Photo) GetPhotoFunc(repository repositories.PhotoRepository) photos.GetPhotoHandlerFunc {
	return func(s photos.GetPhotoParams) middleware.Responder {
		return middleware.ResponderFunc(func(w http.ResponseWriter, _ runtime.Producer) {
			ctx := s.HTTPRequest.Context()
			photo, err := repository.PhotoByID(ctx, s.PhotoID)
			if err != nil {
				p.logger.Error("failed to get photo", zap.Error(err))
				if err := writeErrorInResponse(w, err); err != nil {
					p.logger.Error("failed to response to client", zap.Error(err))
				}
				return
			}
			fileBytes, err := readPhotoFile(photo.ID)
			if err != nil {
				p.logger.Error("failed to read photo file", zap.Error(err))
				if err := writeErrorInResponse(w, err); err != nil {
					p.logger.Error("failed to response to client", zap.Error(err))
				}
				return
			}
			w.Header().Set("Content-Type", "image/jpg")
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(fileBytes)
			if err != nil {
				p.logger.Error("error while writing file", zap.Error(err))
			}
		})
	}
}

func (p Photo) DownloadPhotoFunc(repository repositories.PhotoRepository) photos.DownloadPhotoHandlerFunc {
	return func(s photos.DownloadPhotoParams) middleware.Responder {
		return middleware.ResponderFunc(func(w http.ResponseWriter, _ runtime.Producer) {

			ctx := s.HTTPRequest.Context()
			photo, err := repository.PhotoByID(ctx, s.PhotoID)
			if err != nil {
				p.logger.Error("failed to get photo", zap.Error(err))
				if err := writeErrorInResponse(w, err); err != nil {
					p.logger.Error("failed to response to client", zap.Error(err))
				}
				return
			}
			fileBytes, err := readPhotoFile(photo.ID)
			if err != nil {
				p.logger.Error("failed to read photo file", zap.Error(err))
				if err := writeErrorInResponse(w, err); err != nil {
					p.logger.Error("failed to response to client", zap.Error(err))
				}
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s.jpg", photo.ID))
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(fileBytes)
			if err != nil {
				p.logger.Error("error while writing file", zap.Error(err))
			}
		})
	}
}

func (p Photo) DeletePhotoFunc(repository repositories.PhotoRepository) photos.DeletePhotoHandlerFunc {
	return func(s photos.DeletePhotoParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		photo, err := repository.PhotoByID(ctx, s.PhotoID)
		if err != nil {
			p.logger.Error("failed to get photo", zap.Error(err))
			return photos.NewDeletePhotoDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}

		err = repository.DeletePhotoByID(ctx, photo.ID)
		if err != nil {
			p.logger.Error("delete photo failed", zap.Error(err))
			return photos.NewDeletePhotoDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}
		return photos.NewDeletePhotoOK().WithPayload(&models.DeletePhotoResponse{
			Data: &models.Photo{
				ID:  photo.ID,
				URL: &(photo.URL),
			},
		})
	}
}

func generatePhotoName() (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	name := strings.Replace(id.String(), "-", "", -1)
	return name, nil
}

func savePhotoFile(file io.ReadCloser, name string) error {
	if _, err := os.Stat(photosFolder); os.IsNotExist(err) {
		err = os.Mkdir(photosFolder, 0766)
		if err != nil {
			return err
		}
	}

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(
		filepath.Join(photosFolder, fmt.Sprintf("%s.jpg", name)),
		fileBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func getPhotoURL(name string) (string, error) {
	u, err := url.Parse(photoServerURL)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, photoURLPath, name)
	return u.String(), nil
}

func readPhotoFile(name string) ([]byte, error) {
	fileBytes, err := ioutil.ReadFile(filepath.Join(photosFolder, fmt.Sprintf("%s.jpg", name)))
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}

func deletePhotoFile(name string) error {
	err := os.Remove(filepath.Join(photosFolder, fmt.Sprintf("%s.jpg", name)))
	if err != nil {
		return err
	}
	return nil
}

func writeErrorInResponse(w http.ResponseWriter, err error) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	return json.NewEncoder(w).Encode(buildErrorPayload(err))
}
