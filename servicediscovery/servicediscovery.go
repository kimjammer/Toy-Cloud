package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"kimjammer.com/toycloud/common"
	"net/http"
	"time"
)

const heartbeatInterval = 2

// TODO: Probably is not concurrency safe
var trackedHosts = map[string]common.Heartbeat{}

func main() {
	router := gin.Default()
	router.POST("/newhost", registerNewHost)
	router.GET("/hosts", getLiveWebServices)

	fmt.Println("Service Discovery")

	//Heartbeat all hosts every 5 seconds
	ticker := time.NewTicker(heartbeatInterval * time.Second)
	stopTicker := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				heartbeatHosts()
			case <-stopTicker:
				ticker.Stop()
				return
			}
		}
	}()

	router.Run(":" + common.Port)
}

var respChan = make(chan common.Heartbeat)

func heartbeatHosts() {
	//Send heartbeats
	for _, host := range trackedHosts {
		go sendHeartbeat(host.Address)
	}

	//Receive responses
	for range trackedHosts {
		heartbeat := <-respChan

		//Check if heartbeating host is currently tracked (it should be)
		_, ok := trackedHosts[heartbeat.Address]
		if !ok {
			fmt.Println("Unexpected heartbeat:", heartbeat.Address)
			continue
		}

		//Update host status
		if heartbeat.Success {
			trackedHosts[heartbeat.Address] = heartbeat
		} else {
			newHeartbeat := trackedHosts[heartbeat.Address]
			newHeartbeat.Success = false
			trackedHosts[heartbeat.Address] = newHeartbeat
		}
	}
}

func sendHeartbeat(host string) {
	heartbeat := common.Heartbeat{}
	heartbeat.Address = host

	//Get heartbeat
	resp, err := http.Get("http://" + host + ":" + common.Port + "/heartbeat")
	if err != nil {
		fmt.Println("Heartbeat request failed:", err)
		heartbeat.Success = false
		respChan <- heartbeat
		return
	}

	//Decode response
	err = json.NewDecoder(resp.Body).Decode(&heartbeat)
	if err != nil {
		fmt.Println("Error parsing heartbeat:", err)
		heartbeat.Success = false
		respChan <- heartbeat
		return
	}

	respChan <- heartbeat
}

// Receive from Orchestrator/Containers the IPs of hosts to start tracking
// So we know if a new host started successfully or not
func registerNewHost(c *gin.Context) {
	var hostlist common.HostList

	err := c.ShouldBindJSON(&hostlist)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	for _, host := range hostlist.Hosts {
		_, ok := trackedHosts[host.Address]
		if !ok {
			fmt.Println("Untracked host added:", host)
			trackedHosts[host.Address] = common.Heartbeat{host.Address, 0, 0, false, common.WebService}
		} else {
			fmt.Println("Host already tracked:", host)
		}
	}
}

// Return a list of current live webservice hosts
func getLiveWebServices(c *gin.Context) {
	hostList := []common.Host{}

	for _, host := range trackedHosts {
		if host.Success && host.ServiceType == common.WebService {
			hostList = append(hostList, common.Host{host.Address, host.Load})
		}
	}

	c.JSON(http.StatusOK, common.HostList{hostList})
}
