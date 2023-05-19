package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/therox/modis"
	"golang.org/x/exp/rand"
)

func main() {
	t := time.Now()

	c, err := modis.NewClient(os.Getenv("MODIS_TOKEN"))
	if err != nil {
		fmt.Printf("Error in initializing client: %s", err)
	}

	// modisScenes, err := c.SearchModisScenes(time.Now().Add(-time.Hour*24*1), time.Now())
	modisScenes, err := c.SearchModisScenes(time.Now().Add(-time.Hour*24*20), time.Now().Add(-time.Hour*24*18))
	if err != nil {
		fmt.Printf("Error getting modis scenes:\n%v", err)
	}

	fmt.Printf("Got %d scenes in %s\n", len(modisScenes), time.Since(t))
	t = time.Now()
	ms := modisScenes[rand.Int63n(int64(len(modisScenes)))]

	mss, err := c.GetHDF(ms.ID, ms.Platform, ms.IsArchive)
	if err != nil {
		fmt.Printf("Error getting modis scene:\n%v", err)
	}
	fmt.Printf("Got %s in %s\n", ms.GranuleID, time.Since(t))
	bs, err := io.ReadAll(mss)
	if err != nil {
		fmt.Printf("Error reading modis file:\n%v", err)
	}
	fmt.Println(len(bs))
}
