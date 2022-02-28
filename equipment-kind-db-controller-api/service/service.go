package service

import (
	"net/http"
	"strconv"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/equipment-kind-db-controller-api/model"
	"github.com/gin-gonic/gin"
)

type Service interface {
	GetAllKinds(c *gin.Context) []model.Equipment
	CreateNewKind(c *gin.Context) model.Equipment
	GetKindByID(c *gin.Context)
	DeleteKindByID(c *gin.Context)
}

type service struct {
	equipment []model.Equipment
}

func New() Service {
	return &service{
		equipment: []model.Equipment{},
	}
}

func (s *service) GetAllKinds(c *gin.Context) []model.Equipment {
	return s.equipment
}

func (s *service) GetKindByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	for _, v := range s.equipment {
		if id == v.Id {
			c.IndentedJSON(http.StatusOK, v)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "equipment not found"})
}

func (s *service) CreateNewKind(c *gin.Context) model.Equipment {
	var kind model.Equipment
	if err := c.BindJSON(&kind); err != nil {
		c.IndentedJSON(http.StatusBadRequest, "bad request")
		return model.Equipment{}
	}
	s.equipment = append(s.equipment, kind)
	return kind
}

func (s *service) DeleteKindByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	for i := range s.equipment {
		if id == s.equipment[i].Id {
			s.equipment = append(s.equipment[:i], s.equipment[i+1:]...)
			c.IndentedJSON(http.StatusOK, s.equipment)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "product not found"})
}
