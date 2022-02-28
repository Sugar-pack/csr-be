package main

import (
	"net/http"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/equipment-kind-db-controller-api/service"
	"github.com/gin-gonic/gin"
)

var s service.Service = service.New()

func main() {
	server := gin.Default()
	server.POST("/equipment/kinds", func(c *gin.Context) {
		c.JSON(http.StatusOK, s.CreateNewKind(c))
	})
	server.GET("/equipment/kinds", func(c *gin.Context) {
		c.JSON(http.StatusOK, s.GetAllKinds(c))
	})
	server.GET("/equipment/kinds/:id", s.GetKindByID)
	server.DELETE("/equipment/kinds/:id", s.DeleteKindByID)

	server.Run(":8080")

}
