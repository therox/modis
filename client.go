package modis

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	timefmt "github.com/itchyny/timefmt-go"
	log "github.com/sirupsen/logrus"
)

type storageType string

const (
	storageTypeArchive storageType = "archive"
	storageTypeNRT     storageType = "nrt"
)

type Platform int

const (
	PlatformTerra Platform = iota + 1
	PlatformAqua
)

type ErrScene struct {
	url string
	err error
}

func (e ErrScene) Error() string {
	if e.err != nil {
		return e.url + ": " + e.err.Error()
	}
	return e.url + ": <nil>"
}

type ErrScenes struct {
	Errors []error
}

func (e *ErrScenes) Error() string {
	s := strings.Builder{}
	for _, err := range e.Errors {
		if err != nil {
			s.WriteString(err.Error() + "\n")
			continue
		}
		s.WriteString(fmt.Sprintf("%v\n", err))
	}
	return s.String()
}

type Client struct {
	modisToken   string
	httpClient   *http.Client
	nrtValidDays int
}

func NewClient(modisToken string) (*Client, error) {
	if modisToken == "" {
		return nil, errors.New("modisToken is empty")
	}

	log.SetLevel(log.DebugLevel)

	return &Client{
		modisToken:   modisToken,
		httpClient:   &http.Client{},
		nrtValidDays: 8,
	}, nil
}

func (c *Client) GetModisScenes(startDate time.Time, endDate time.Time) (ModisData, error) {
	if endDate.After(time.Now()) {
		endDate = time.Now()
	}
	var msErr = new(ErrScenes)
	if startDate.After(endDate) {
		return nil, errors.New("startDate is after endDate")
	}
	log.Debug("Searching for MODIS product...")

	// get scenes from MODIS
	res := ModisData{}
	noErrCount := 0
	for _, platformType := range []Platform{PlatformTerra, PlatformAqua} {
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {

			// For dates before now() - nrtValidDays, use archive data
			storageType := storageTypeArchive
			if d.After(time.Now().AddDate(0, 0, -c.nrtValidDays)) {
				// For dates after now() - nrtValidDays, use near real time data
				storageType = storageTypeNRT
			}

			dlURL := fmt.Sprintf("%s%s", platformURLs[storageType]["meta"][platformType]["url"], timefmt.Format(d, platformURLs[storageType]["meta"][platformType]["txt_template"]))
			log.Debugf("Downloading %v from %s", platformType, dlURL)
			content, err := c.getData(dlURL, false)
			if err != nil {
				msErr.Errors = append(msErr.Errors, ErrScene{dlURL, fmt.Errorf("error getting data: %s", err)})
				continue
			}
			mdScenes, err := parseModisData(content, platformType, storageType == "archive")
			if err != nil {
				msErr.Errors = append(msErr.Errors, ErrScene{dlURL, fmt.Errorf("error parsing data: %s", err)})
				continue
			}
			res = append(res, mdScenes...)
			msErr.Errors = append(msErr.Errors, nil)
			noErrCount++
		}
	}
	if len(msErr.Errors) != noErrCount {
		return res, msErr
	}
	return res, nil
}

func (c *Client) getData(downloadURL string, download bool) ([]byte, error) {
	fmt.Println("Downloading from ", downloadURL)
	req, err := http.NewRequest("GET", downloadURL, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.modisToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if download {
		// get filename from url - this is last part of url
		filename := filepath.Base(downloadURL)
		// write to file
		err = os.WriteFile(filename, bs, 0644)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	return bs, nil
}

type ParseErrors []error

func (e ParseErrors) Error() string {
	s := strings.Builder{}
	for _, err := range e {
		s.WriteString(err.Error() + "\n")
	}
	return s.String()
}

func parseModisData(data []byte, platformTypeID Platform, isArchive bool) (ModisData, error) {
	res := make([]ModisDataItem, 0)
	var pErrors ParseErrors
P:
	for i, line := range strings.Split(string(data), "\n") {
		fmt.Printf("[%d] > %+v\n", i+1, line)
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			// if strings.TrimSpace(line) == "" {
			// skip header or comment
			continue
		}
		s := strings.Split(line, ",")
		if len(s) != 17 {
			pErrors = append(pErrors, fmt.Errorf("line %d has %d fields, expected 17, skipping", i+1, len(s)))
			continue
		}
		mdItem := ModisDataItem{
			GranuleID: s[0],
		}

		sd, err := time.Parse("2006-01-02 15:04", s[1])
		if err != nil {
			pErrors = append(pErrors, fmt.Errorf("line %d has invalid date %q, skipping", i+1, s[1]))
			continue
		}
		mdItem.StartDateTime = sd

		as, err := strconv.Atoi(s[2])
		if err != nil {
			pErrors = append(pErrors, fmt.Errorf("line %d has invalid archive set, skipping", i+1))
			continue
		}
		mdItem.ArchiveSet = as

		orbit, err := strconv.ParseInt(s[3], 10, 64)
		if err != nil {
			pErrors = append(pErrors, fmt.Errorf("line %d has invalid orbit, skipping", i+1))
			continue
		}
		mdItem.OrbitNumber = orbit

		if s[4] != "N" && s[4] != "D" && s[4] != "B" {
			pErrors = append(pErrors, fmt.Errorf("line %d has invalid day/night flag %s, skipping", i+1, s[4]))
			continue
		}

		mdItem.DayNightFlag = s[4]

		// Bounding box
		mdItem.EastBoundingCoord, err = strconv.ParseFloat(s[5], 64)
		if err != nil {
			pErrors = append(pErrors, fmt.Errorf("line %d has invalid east bounding coord, skipping", i+1))
			continue
		}
		mdItem.NorthBoundingCoord, err = strconv.ParseFloat(s[6], 64)
		if err != nil {
			pErrors = append(pErrors, fmt.Errorf("line %d has invalid north bounding coord, skipping", i+1))
			continue
		}

		mdItem.SouthBoundingCoord, err = strconv.ParseFloat(s[7], 64)
		if err != nil {
			pErrors = append(pErrors, fmt.Errorf("line %d has invalid south bounding coord, skipping", i+1))
			continue
		}

		mdItem.WestBoundingCoord, err = strconv.ParseFloat(s[8], 64)
		if err != nil {
			pErrors = append(pErrors, fmt.Errorf("line %d has invalid west bounding coord, skipping", i+1))
			continue
		}

		lats := []float64{}
		// lats - last 4 fields from s in reversed order
		for _, lat := range s[len(s)-4:] {
			latf, err := strconv.ParseFloat(lat, 64)
			if err != nil {
				pErrors = append(pErrors, fmt.Errorf("line %d has invalid lat, skipping", i+1))
				continue P
			}
			lats = append(lats, latf)
		}
		// reverse lats
		for i, j := 0, len(lats)-1; i < j; i, j = i+1, j-1 {
			lats[i], lats[j] = lats[j], lats[i]
		}

		lons := []float64{}
		// lons - last 8 fields from s in reversed order
		for _, lon := range s[len(s)-8 : len(s)-4] {
			lonf, err := strconv.ParseFloat(lon, 64)
			if err != nil {
				pErrors = append(pErrors, fmt.Errorf("line %d has invalid lon, skipping", i+1))
				continue P
			}
			lons = append(lons, lonf)
		}
		// reverse lons
		for i, j := 0, len(lons)-1; i < j; i, j = i+1, j-1 {
			lons[i], lons[j] = lons[j], lons[i]
		}

		lats = append(lats, lats[0])
		lons = append(lons, lons[0])

		mdItem.ShapeWKT = fmt.Sprintf("POLYGON((%s))", strings.Join(wktCoordsWithPrecision(lons, lats, 13), ","))
		mdItem.IsArchive = isArchive

		res = append(res, mdItem)
	}

	if len(pErrors) > 0 {
		return res, pErrors
	}

	return res, nil
}

func wktCoordsWithPrecision(lons []float64, lats []float64, precision int) []string {
	res := []string{}
	for i := range lons {
		res = append(res, fmt.Sprintf("%.*f %.*f", precision, lons[i], precision, lats[i]))
	}
	return res
}
