package modis

const (
	PlatformTerra = iota + 1
	PlatformAqua
)

var (
	platformURLs = map[storageType]map[string]map[int]map[string]string{
		"nrt": {
			"meta": {
				PlatformTerra: {
					"url":          "https://nrt4.modaps.eosdis.nasa.gov/api/v2/content/archives/geoMetaMODIS/61/TERRA/",
					"txt_template": "%Y/MOD03_%Y-%m-%d.txt",
				},
				PlatformAqua: {
					"url":          "https://nrt4.modaps.eosdis.nasa.gov/api/v2/content/archives/geoMetaMODIS/61/AQUA/",
					"txt_template": "%Y/MYD03_%Y-%m-%d.txt",
				},
			},
		},
		"archive": {
			"meta": {
				PlatformTerra: {
					"url":          "https://ladsweb.modaps.eosdis.nasa.gov/archive/geoMeta/61/TERRA/",
					"txt_template": "%Y/MOD03_%Y-%m-%d.txt",
				},
				PlatformAqua: {
					"url":          "https://ladsweb.modaps.eosdis.nasa.gov/archive/geoMeta/61/AQUA/",
					"txt_template": "%Y/MYD03_%Y-%m-%d.txt",
				},
			},
		},
	}
)
