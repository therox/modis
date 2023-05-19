package modis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type FileInfo struct {
	Content []struct {
		ArchiveSets   int    `json:"archiveSets"`
		Cksum         string `json:"cksum"`
		DataDay       string `json:"dataDay"`
		DownloadsLink string `json:"downloadsLink"`
		FileID        int64  `json:"fileId"`
		Md5Sum        string `json:"md5sum"`
		Mtime         int    `json:"mtime"`
		Name          string `json:"name"`
		Products      string `json:"products"`
		ResourceType  string `json:"resourceType"`
		Self          string `json:"self"`
		Size          int    `json:"size"`
	} `json:"content"`
	DownloadsLink string `json:"downloadsLink"`
	FileCount     int    `json:"file_count"`
	Mtime         int    `json:"mtime"`
	Name          string `json:"name"`
	ResourceType  string `json:"resourceType"`
	Self          string `json:"self"`
	Size          int    `json:"size"`
}

func (c *Client) GetHDF(sceneID string, scenePlatform Platform, isArchive bool) (io.Reader, error) {

	repository := storageTypeArchive
	if !isArchive {
		repository = storageTypeNRT
	}

	nOfDay := string(sceneID[len(sceneID)-3:])
	sceneYear := string(sceneID[1 : len(sceneID)-3])

	hdfURL := ""

	if isArchive {
		// Get json file for actual data
		jsonURL := fmt.Sprintf("%s%s/%s.json", platformURLs[repository]["meta"][scenePlatform]["file_url"], sceneYear, nOfDay)
		r, err := c.getContent(jsonURL)
		if err != nil {
			return nil, fmt.Errorf("Error loading json file %s for modis scene id %s and product type %d. Error: %w", jsonURL, sceneID, scenePlatform, err)
		}
		bs, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		var fi FileInfo
		err = json.Unmarshal(bs, &fi)
		if err != nil {
			return nil, err
		}

		// Search for file with same name as granule id
		isFound := false
		for _, f := range fi.Content {
			if strings.HasPrefix(f.Name, fmt.Sprintf("%s.%s.061", platformURLs[repository]["meta"][scenePlatform]["file_prefix"], sceneID)) {
				isFound = true
				hdfURL = f.DownloadsLink
				break
			}
		}

		// err_txt = (
		// 	f"Error loading hdf file for modis scene id {modis_scene.sceneid} and product type"
		// 	f" {modis_scene.modis_product_type}. Filename not found in json file {json_url}"
		// )
		if !isFound {
			return nil, fmt.Errorf("Error loading hdf file for modis scene id %s and product type %d. Filename not found in json file %s", sceneID, scenePlatform, jsonURL)
		}

		// bs, _ = json.MarshalIndent(fi, "", "  ")
		// fmt.Println(string(bs))

	} else {
		// hdf_url = (
		// 	self._product_urls["nrt"][modis_scene.modis_product_type]["url"]
		// 	+ f"{modis_scene.scene_datetime.year}/{n_of_day}"
		// 	f"/{self._product_urls['nrt'][modis_scene.modis_product_type]['file_prefix']}.{modis_scene.sceneid}.061.NRT.hdf"
		// )
		hdfURL = fmt.Sprintf("%s%s/%s/%s.%s.061.NRT.hdf", platformURLs[repository]["meta"][scenePlatform]["file_url"], sceneYear, nOfDay, platformURLs[repository]["meta"][scenePlatform]["file_prefix"], sceneID)
	}

	req, err := http.NewRequest("GET", hdfURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.modisToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := bytes.NewReader(bs)
	return res, nil
}

func (c *Client) getContent(fileURL string) (io.Reader, error) {
	req, err := http.NewRequest(http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.modisToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(bs), nil
}
