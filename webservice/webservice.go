package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func main() {
	router := gin.Default()
	router.GET("/ping", ping)

	router.Run(":8080")
}

func ping(c *gin.Context) {
	hostname, _ := os.Hostname()
	c.IndentedJSON(http.StatusOK, gin.H{"message": "pong from host: " + hostname})
}
