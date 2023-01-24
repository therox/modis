package modis

import "time"

type ModisScene struct {
	ID                  string
	Datetime            time.Time
	Contour             string
	ModisPlatformTypeID int
	IsArchive           bool
	ToLoad              bool
}

type ModisDataItem struct {
	GranuleID          string
	StartDateTime      time.Time
	ArchiveSet         int
	OrbitNumber        int64
	DayNightFlag       string
	EastBoundingCoord  float64
	NorthBoundingCoord float64
	SouthBoundingCoord float64
	WestBoundingCoord  float64
	ShapeWKT           string
	IsArchive          bool
}

type ModisData []ModisDataItem
