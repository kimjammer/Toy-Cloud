package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"kimjammer.com/toycloud/common"
	"net/http"
	"os"
	"sync"
	"time"
)

const registrationRetryTime = 2

type loadMetrics struct {
	requests           int //Unused
	processingDuration time.Duration
	resetTime          time.Time
	mu                 sync.Mutex
}

var load = loadMetrics{0, 0, time.Now(), sync.Mutex{}}

func main() {
	router := gin.Default()
	router.GET("/ping", hostLoadMiddleware(), ping)
	router.GET("/heartbeat", handleHeartbeat)

	//Register with service discovery
	go registerService()

	router.Run(":" + common.Port)
}

func hostLoadMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		load.mu.Lock()
		defer load.mu.Unlock()
		load.processingDuration += time.Since(start)
		load.requests++
	}
}

// Returns the percent of time spent handling requests since it was last read
// Interval at which this is reset(read) depends on the heartbeat interval
func crrLoad() float64 {
	load.mu.Lock()
	fmt.Println("ProcessingDuration:", load.processingDuration)
	fmt.Println("Requests:", load.requests)
	busyPercent := load.processingDuration.Seconds() / time.Since(load.resetTime).Seconds()
	fmt.Println("BusyPercent:", busyPercent)
	load.requests = 0
	load.processingDuration = 0
	load.resetTime = time.Now()
	load.mu.Unlock()

	return busyPercent
}

func ping(c *gin.Context) {
	hostname, err := os.Hostname()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "pong from host: " + hostname})
}

func handleHeartbeat(c *gin.Context) {
	//Get hostname (works like IP)
	hostname, err := os.Hostname()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	heartbeat := common.Heartbeat{hostname, time.Now().Unix(), crrLoad(), true, common.WebService}
	c.IndentedJSON(http.StatusOK, heartbeat)
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
