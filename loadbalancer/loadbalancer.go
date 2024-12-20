package main

import (
	"bufio"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/url"
)

type Host struct {
	address string
	alive   bool
	load    float64
}

var registeredHosts = []Host{
	{"http://webservice-1:8080", true, 0},
	{"http://webservice-2:8080", true, 0},
}

// Returns a function that chooses hosts
// Chooses hosts in a round robin style, rotating hosts
func roundRobin() func() string {
	hostIndex := 0

	var chooser = func() string {
		numHosts := len(registeredHosts)
		if hostIndex < numHosts-1 {
			hostIndex++
		} else {
			hostIndex = 0
		}
		return registeredHosts[hostIndex].address
	}
	return chooser
}

var strategy = roundRobin()

func getLoadBalancedHost() string {
	return strategy()
}

func main() {
	r := gin.Default()
	r.GET("/:path", func(c *gin.Context) {
		//Resolve host address, change scheme and host for request
		req := c.Request
		proxy, err := url.Parse(getLoadBalancedHost())
		if err != nil {
			log.Printf("Error in parsing host address: %v", err)
			c.String(500, "error")
			return
		}
		req.URL.Scheme = proxy.Scheme
		req.URL.Host = proxy.Host

		//Get actual request from host
		transport := http.DefaultTransport
		resp, err := transport.RoundTrip(req)
		if err != nil {
			log.Printf("Error in roundtrip: %v", err)
			c.String(500, "error")
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
	})

	r.Run(":8080")
}
