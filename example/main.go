package main

import (
	"fmt"
	"os"
	"time"

	"github.com/therox/modis"
)

func main() {
	modisToken := os.Getenv("MODIS_TOKEN")
	c, err := modis.NewClient(modisToken)
	if err != nil {
		fmt.Printf("Error in initializing client: %s", err)
	}

	modisScenes, err := c.GetModisScenes(time.Now().Add(-time.Hour*24*20), time.Now())
	if err != nil {
		fmt.Printf("Error getting modis scenes:\n%s", err)
	}

	fmt.Printf("Got %d scenes", len(modisScenes))
}
