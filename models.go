package modis

import (
	"errors"
	"strings"
	"time"
)

// type ModisScene struct {
// 	ID                  string
// 	Datetime            time.Time
// 	Contour             string
// 	ModisPlatformTypeID Platform
// 	IsArchive           bool
// 	ToLoad              bool
// }

type ModisScene struct {
	ID                 string
	GranuleID          string
	StartDateTime      time.Time
	ArchiveSet         int
	OrbitNumber        int64
	DayNightFlag       string
	EastBoundingCoord  float64
	NorthBoundingCoord float64
	SouthBoundingCoord float64
	WestBoundingCoord  float64
	Platform           Platform
	ShapeWKT           string
	IsArchive          bool
}

// func (mdi ModisDataItem) ModisScene() ModisScene {
// 	p, _ := platform(mdi.GranuleID)
// 	return ModisScene{
// 		ID:                  strings.Join(strings.Split(mdi.GranuleID, ".")[:2], "."),
// 		Datetime:            mdi.StartDateTime,
// 		Contour:             mdi.ShapeWKT,
// 		ModisPlatformTypeID: p,
// 		IsArchive:           mdi.IsArchive,
// 	}
// }

func platform(id string) (Platform, error) {
	// Get platform from granule id
	switch strings.Split(id, ".")[0] {
	case "MOD03":
		return PlatformTerra, nil
	case "MYD03":
		return PlatformAqua, nil
	}
	return PlatformTerra, errors.New("invalid platform")
}

type ModisData []ModisScene
