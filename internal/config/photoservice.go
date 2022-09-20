package config

import "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"

type PhotoService struct {
	PhotosFolder string
}

func NewPhotoServiceConfig() *PhotoService {
	photosFolder := utils.GetEnv("PHOTOS_FOLDER", "equipments_photos")
	return &PhotoService{
		PhotosFolder: photosFolder,
	}
}
