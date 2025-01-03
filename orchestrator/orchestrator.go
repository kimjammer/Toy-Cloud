package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"kimjammer.com/toycloud/common"
	"math/rand/v2"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	router := gin.Default()
	router.POST("/start-container", startContainer)
	router.POST("/stop-container", stopContainer)
	router.GET("/docker-info", dockerInfo)

	router.Run(":" + common.Port)
}

func startContainer(c *gin.Context) {
	var command common.StartContainerCommand

	err := c.ShouldBindJSON(&command)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	//Run the docker command
	randNum := rand.IntN(999)
	cmd := exec.Command("docker", "run", "--network", "toycloud_default", "--name", command.Image+"-"+strconv.Itoa(randNum), command.Image)
	err = cmd.Run()
	if err != nil {
		fmt.Println("Failed to start container: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func stopContainer(c *gin.Context) {
	var command common.StopContainerCommand

	err := c.ShouldBindJSON(&command)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	cmd := exec.Command("docker", "stop", command.ID)
	err = cmd.Run()
	if err != nil {
		fmt.Println("Failed to stop container: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func dockerInfo(c *gin.Context) {
	cmd := exec.Command("docker", "ps", "-a")
	out, err := cmd.Output()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Failed to get docker info": err.Error()})
	}

	outString := string(out)
	lines := strings.Split(outString, "\n")

	containerList := []common.ContainerStatus{}

	for i, line := range lines {
		//Skip first line with column titles
		if i == 0 {
			continue
		}

		sections := strings.Split(line, "   ")
		container := common.ContainerStatus{
			ContainerID: sections[0],
			Image:       sections[1],
			Command:     sections[2],
			Created:     sections[3],
			Status:      sections[4],
			Ports:       sections[5],
			Names:       sections[6],
		}
		containerList = append(containerList, container)
	}

	dockerStatus := common.DockerStatus{containerList}

	c.JSON(http.StatusOK, dockerStatus)
}
