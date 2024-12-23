package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"kimjammer.com/toycloud/common"
	"net/http"
	"os"
	"time"
)

const registrationRetryTime = 2

func main() {
	router := gin.Default()
	router.GET("/ping", ping)
	router.GET("/heartbeat", handleHeartbeat)

	//Register with service discovery
	go registerService()

	router.Run(":" + common.Port)
}

// TODO: Implement measuring current load
func crrLoad() float64 {
	return 0
}

func ping(c *gin.Context) {
	hostname, err := os.Hostname()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "pong from host: " + hostname})
}

func handleHeartbeat(c *gin.Context) {
	hostname, err := os.Hostname()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	heartbeat := common.Heartbeat{hostname, time.Now().Unix(), crrLoad(), true, common.WebService}
	c.IndentedJSON(http.StatusOK, heartbeat)
	fmt.Println("Heartbeat")
}

func registerService() {
	hostname, _ := os.Hostname()

	hostList := common.HostList{[]common.Host{{hostname, 0}}}
	hostListJSON, _ := json.Marshal(hostList)

	success := false

	for !success {
		_, err := http.Post("http://"+"servicediscovery"+":"+common.Port+"/newhost", "application/json",
			bytes.NewBuffer(hostListJSON))
		if err != nil {
			fmt.Println("Failed to register with service discovery")
			time.Sleep(registrationRetryTime * time.Second)
		} else {
			fmt.Println("Registered with service discovery:", hostname)
			success = true
		}
	}
}
