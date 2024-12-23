package main

import (
	"bufio"
	"encoding/json"
	"errors"
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

// Returns a function that chooses hosts.
// Chooses hosts in a round-robin style, rotating hosts
func roundRobin() func() (string, error) {
	hostIndex := 0

	return func() (string, error) {
		registeredHosts.mu.Lock()
		defer registeredHosts.mu.Unlock()

		numHosts := len(registeredHosts.sl)
		if numHosts == 0 {
			return "", errors.New("No available hosts")
		}

		if hostIndex < numHosts-1 {
			hostIndex++
		} else {
			hostIndex = 0
		}
		return registeredHosts.sl[hostIndex].Address, nil
	}
}

// Returns a function that chooses hosts.
// Chooses the host with the lowest load
func lowestLoad() func() (string, error) {

	return func() (string, error) {
		registeredHosts.mu.Lock()
		defer registeredHosts.mu.Unlock()

		numHosts := len(registeredHosts.sl)
		if numHosts == 0 {
			return "", errors.New("No available hosts")
		}

		minLoadHost := registeredHosts.sl[0].Address
		minLoad := registeredHosts.sl[0].Load
		for _, host := range registeredHosts.sl {
			fmt.Println("Host:", host.Address, "Load:", host.Load)
			if host.Load < minLoad {
				minLoadHost = host.Address
				minLoad = host.Load
			}
		}

		return minLoadHost, nil
	}
}

var strategy = roundRobin()

func getLoadBalancedHost() (string, error) {
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
	host, err := getLoadBalancedHost()
	if err != nil {
		fmt.Println("Error in finding host:", err)
		c.String(http.StatusInternalServerError, "error")
		return
	}

	hosturl, err := url.Parse("http://" + host + ":" + common.Port)
	if err != nil {
		fmt.Println("Error in parsing host address:", err)
		c.String(http.StatusInternalServerError, "error")
		return
	}
	req.URL.Scheme = hosturl.Scheme
	req.URL.Host = hosturl.Host

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

	prevNumHosts := len(registeredHosts.sl)

	registeredHosts.sl = hostList.Hosts

	if prevNumHosts != len(registeredHosts.sl) {
		fmt.Println("Num Hosts:", len(registeredHosts.sl))
	}
}
