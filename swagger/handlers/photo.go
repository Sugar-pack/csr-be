package handlers

import (
	"encoding/json"
	"fmt"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/restapi/operations/photos"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/repositories"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/services"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

const photoURLPath string = "api/equipment/photos/"

type Photo struct {
	serverURL string
	logger    *zap.Logger
}

func NewPhoto(serverURL string, logger *zap.Logger) *Photo {
	return &Photo{
		serverURL: serverURL,
		logger:    logger,
	}
}

func (p Photo) CreateNewPhotoFunc(repository repositories.PhotoRepository,
	fileManager services.FileManager) photos.CreateNewPhotoHandlerFunc {
	return func(s photos.CreateNewPhotoParams) middleware.Responder {
		ctx := s.HTTPRequest.Context()
		// read input file
		fileBytes, err := ioutil.ReadAll(s.File)
		if err != nil {
			p.logger.Error("failed to read file", zap.Error(err))
			return photos.NewCreateNewPhotoDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}
		if err := s.File.Close(); err != nil {
			p.logger.Error("Failed to close file", zap.Error(err))
			return photos.NewCreateNewPhotoDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}
		// check if file is not empty
		if len(fileBytes) == 0 {
			p.logger.Error("file is empty")
			return photos.NewCreateNewPhotoDefault(http.StatusBadRequest).
				WithPayload(buildStringPayload("File is empty"))
		}
		// check if file is image jpg/jpeg
		mimeType := http.DetectContentType(fileBytes)
		if mimeType != "image/jpg" && mimeType != "image/jpeg" {
			p.logger.Error(fmt.Sprintf("wrong file format: %s. file should be jpg or jpeg", mimeType))
			return photos.NewCreateNewPhotoDefault(http.StatusBadRequest).
				WithPayload(buildStringPayload("Wrong file format. File should be jpg or jpeg"))
		}

		photoID, err := fileManager.GenerateFileName()
		if err != nil {
			p.logger.Error("failed to generate photo name", zap.Error(err))
			return photos.NewCreateNewPhotoDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}

		fileName := fmt.Sprintf("%s.jpg", photoID)
		if err := fileManager.SaveDataToFile(fileBytes, fileName); err != nil {
			p.logger.Error("failed to save photo to file", zap.Error(err))
			return photos.NewCreateNewPhotoDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}
		photoURL, err := fileManager.BuildFileURL(p.serverURL, photoURLPath, photoID)
		if err != nil {
			p.logger.Error("failed to get photo url", zap.Error(err))
			return photos.NewCreateNewPhotoDefault(http.StatusInternalServerError).
				WithPayload(buildErrorPayload(err))
		}
		newPhoto := models.Photo{
			ID:       photoID,
			URL:      &photoURL,
			FileName: fileName,
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

func (p Photo) GetPhotoFunc(repository repositories.PhotoRepository,
	fileManager services.FileManager) photos.GetPhotoHandlerFunc {
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
			fileBytes, err := fileManager.ReadFile(photo.FileName)
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

func (p Photo) DownloadPhotoFunc(repository repositories.PhotoRepository,
	fileManager services.FileManager) photos.DownloadPhotoHandlerFunc {
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
			fileBytes, err := fileManager.ReadFile(photo.FileName)
			if err != nil {
				p.logger.Error("failed to read photo file", zap.Error(err))
				if err := writeErrorInResponse(w, err); err != nil {
					p.logger.Error("failed to response to client", zap.Error(err))
				}
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", photo.FileName))
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(fileBytes)
			if err != nil {
				p.logger.Error("error while writing file", zap.Error(err))
			}
		})
	}
}

func (p Photo) DeletePhotoFunc(repository repositories.PhotoRepository,
	fileManager services.FileManager) photos.DeletePhotoHandlerFunc {
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
		err = fileManager.DeleteFile(photo.FileName)
		if err != nil {
			p.logger.Error("failed to delete photo file", zap.Error(err))
		}

		return photos.NewDeletePhotoOK().WithPayload("Photo deleted")
	}
}

func writeErrorInResponse(w http.ResponseWriter, err error) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	return json.NewEncoder(w).Encode(buildErrorPayload(err))
}
