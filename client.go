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

func (c *Client) GetModisScenes(startDate time.Time, endDate time.Time) ([]ModisScene, error) {
	if startDate.After(endDate) {
		return nil, errors.New("startDate is after endDate")
	}
	log.Debug("Searching for MODIS product...")

	// get scenes from MODIS
	res := []ModisScene{}
	for _, platformType := range []int{PlatformTerra, PlatformAqua} {
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {

			// For dates before now() - nrtValidDays, use archive data
			storageType := "archive"
			if d.After(time.Now().AddDate(0, 0, -c.nrtValidDays)) {
				// For dates after now() - nrtValidDays, use near real time data
				storageType = "nrt"
			}

			dlURL := fmt.Sprintf("%s%s", platformURLs[storageType]["meta"][platformType]["url"], timefmt.Format(d, platformURLs[storageType]["meta"][platformType]["txt_template"]))
			log.Debugf("Downloading %v from %s", platformType, dlURL)
			content, err := c.getData(dlURL, false)
			if err != nil {
				continue
			}
			mdScenes, err := parseModisData(content, platformType, false)
			if err != nil {
				return nil, err
			}
			res = append(res, mdScenes...)
		}
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

func parseModisData(data []byte, platformTypeID int, isArchive bool) ([]ModisScene, error) {
	res := []ModisScene{}
P:
	for i, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		s := strings.Split(line, ",")
		if len(s) != 17 {
			log.Warningf("Line %d has %d fields, expected 17. Skipping.", i, len(s))
			continue
		}
		sd, err := time.Parse("2006-01-02 15:04", s[1])
		if err != nil {
			log.Warningf("Line %d has invalid date %q. Skipping.", i, s[1])
			continue
		}
		// as, err := strconv.Atoi(s[2])
		// if err != nil {
		// 	log.Warning("Line %d has invalid archive set. Skipping.", i)
		// 	continue
		// }
		// orbit, err := strconv.Atoi(s[3])
		// if err != nil {
		// 	log.Warning("Line %d has invalid orbit. Skipping.", i)
		// 	continue
		// }

		granuleID := s[0]
		startDate := sd
		// archiveSet := as
		// orbit := orbit
		dayNightFlag := s[4]

		//  lats = list(map(float, ln.split(",")[-4::]))[::-1]
		// lons = list(map(float, ln.split(",")[-8:-4]))[::-1]
		// lons.append(lons[0])
		// lats.append(lats[0])

		lats := []float64{}
		// lats - last 4 fields from s in reversed order
		for _, lat := range s[len(s)-4:] {
			latf, err := strconv.ParseFloat(lat, 64)
			if err != nil {
				log.Warningf("Line %d has invalid lat. Skipping.", i)
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
				log.Warningf("Line %d has invalid lon. Skipping.", i)
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

		if Mean(lats) > 0 {
			if dayNightFlag != "D" {
				continue P
			}

			// Make wkt from lons and lats with presision of 13 decimal places
			wkt := fmt.Sprintf("POLYGON((%s))", strings.Join(wktCoordsWithPrecision(lons, lats, 13), ","))

			ms := ModisScene{
				ID:                  granuleID[6:19],
				Datetime:            startDate,
				Contour:             wkt,
				ModisPlatformTypeID: platformTypeID,
				IsArchive:           isArchive,
				ToLoad:              false,
			}

			res = append(res, ms)
		}
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
