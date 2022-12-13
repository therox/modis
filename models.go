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
