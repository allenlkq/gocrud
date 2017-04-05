package main

import (
	"gopkg.in/gin-gonic/gin.v1"
	"net/http"
	. "github.com/manishrjain/gocrud/entities"
	"io/ioutil"
	"strings"
	"path/filepath"
	"reflect"
)

const (
	pkg = "entities"
	tmpDir = "/tmp/test"
	httpPort = ":8090"
)

var entities []string

func main() {
	router := gin.Default()

	// get all entities
	files, _ := ioutil.ReadDir("./" + pkg)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		fileName := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		entities = append(entities, fileName)
	}

	router.POST("/entity/:name", createEntity)
	router.Run(httpPort)
}

func createEntity(c *gin.Context) {
	name := c.Param("name")
	if !stringInSlice(name, entities) {
		c.JSON(http.StatusNotFound, gin.H{"status": "failed"})
	}
	if name == "structure" {
		var json Structure
		if c.BindJSON(&json) == nil {
			createStructure(json)
			c.JSON(http.StatusOK, gin.H{"status": "successful"})
		}
		c.JSON(http.StatusBadRequest, gin.H{"status": "failed"})
	}
	var json map[string]string
	if c.BindJSON(&json) == nil {
		new
		c.JSON(http.StatusOK, gin.H{"status": "successful"})
	}
	c.JSON(http.StatusBadRequest, gin.H{"status": "failed"})
}
