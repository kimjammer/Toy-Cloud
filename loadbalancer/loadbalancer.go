package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"kimjammer.com/toycloud/common"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const hostUpdateInterval = 1

type safeHostList struct {
	mu sync.Mutex
	sl []common.Host
}

var registeredHosts = safeHostList{sl: make([]common.Host, 0)}

// Returns a function that chooses hosts
// Chooses hosts in a round robin style, rotating hosts
func roundRobin() func() string {
	hostIndex := 0
	registeredHosts.mu.Lock()
	defer registeredHosts.mu.Unlock()

	var chooser = func() string {
		numHosts := len(registeredHosts.sl)
		if hostIndex < numHosts-1 {
			hostIndex++
		} else {
			hostIndex = 0
		}
		return "http://" + registeredHosts.sl[hostIndex].Address + ":" + common.Port
	}
	return chooser
}

var strategy = roundRobin()

func getLoadBalancedHost() string {
	return strategy()
}

func main() {
	router := gin.Default()
	router.GET("/:path", handleRequest)

	//Get updated hosts
	ticker := time.NewTicker(hostUpdateInterval * time.Second)
	stopTicker := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				getHosts()
			case <-stopTicker:
				ticker.Stop()
				return
			}
		}
	}()

	router.Run(":" + common.Port)
}

func handleRequest(c *gin.Context) {
	//Resolve host address, change scheme and host for request
	req := c.Request
	proxy, err := url.Parse(getLoadBalancedHost())
	if err != nil {
		fmt.Println("Error in parsing host address:", err)
		c.String(http.StatusInternalServerError, "error")
		return
	}
	req.URL.Scheme = proxy.Scheme
	req.URL.Host = proxy.Host

	//Get actual request from host
	transport := http.DefaultTransport
	resp, err := transport.RoundTrip(req)
	if err != nil {
		fmt.Println("Error in roundtrip:", err)
		c.String(http.StatusInternalServerError, "error")
		return
	}

	//Return response upstream
	for k, vv := range resp.Header {
		for _, v := range vv {
			c.Header(k, v)
		}
	}
	defer resp.Body.Close()
	bufio.NewReader(resp.Body).WriteTo(c.Writer)
	return
}

func getHosts() {
	hostList := common.HostList{}

	//Get live host list
	resp, err := http.Get("http://" + "servicediscovery" + ":" + common.Port + "/hosts")
	if err != nil {
		fmt.Println("Fetching host list failed:", err)
		return
	}

	//Decode response
	err = json.NewDecoder(resp.Body).Decode(&hostList)
	if err != nil {
		fmt.Println("Error parsing host list:", err)
		return
	}

	registeredHosts.mu.Lock()
	defer registeredHosts.mu.Unlock()

	registeredHosts.sl = hostList.Hosts
	fmt.Println("Num Hosts:", len(registeredHosts.sl))
}
