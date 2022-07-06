package config

import "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"

type PhotoService struct {
	PhotosFolder    string
	PhotosServerURL string
}

func NewPhotoServiceConfig() *PhotoService {
	photosServerURL := utils.GetEnv("PHOTOS_SERVER_URL", "http://localhost:8080/")
	photosFolder := utils.GetEnv("PHOTOS_FOLDER", "equipments_photos")
	return &PhotoService{
		PhotosServerURL: photosServerURL,
		PhotosFolder:    photosFolder,
	}
}
