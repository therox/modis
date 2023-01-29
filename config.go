package modis

var (
	platformURLs = map[storageType]map[string]map[Platform]map[string]string{
		"nrt": {
			"meta": {
				PlatformTerra: {
					"url":          "https://nrt4.modaps.eosdis.nasa.gov/api/v2/content/archives/geoMetaMODIS/61/TERRA/",
					"txt_template": "%Y/MOD03_%Y-%m-%d.txt",
					"file_url":     "https://nrt4.modaps.eosdis.nasa.gov/api/v2/content/archives/allData/61/MOD09/",
					"file_prefix":  "MOD09",
				},
				PlatformAqua: {
					"url":          "https://nrt4.modaps.eosdis.nasa.gov/api/v2/content/archives/geoMetaMODIS/61/AQUA/",
					"txt_template": "%Y/MYD03_%Y-%m-%d.txt",
					"file_url":     "https://nrt4.modaps.eosdis.nasa.gov/api/v2/content/archives/allData/61/MYD09/",
					"file_prefix":  "MYD09",
				},
			},
		},
		"archive": {
			"meta": {
				PlatformTerra: {
					"url":          "https://ladsweb.modaps.eosdis.nasa.gov/archive/geoMeta/61/TERRA/",
					"txt_template": "%Y/MOD03_%Y-%m-%d.txt",
					"file_url":     "https://ladsweb.modaps.eosdis.nasa.gov/archive/allData/61/MOD09/",
					"file_prefix":  "MOD09",
				},
				PlatformAqua: {
					"url":          "https://ladsweb.modaps.eosdis.nasa.gov/archive/geoMeta/61/AQUA/",
					"txt_template": "%Y/MYD03_%Y-%m-%d.txt",
					"file_url":     "https://ladsweb.modaps.eosdis.nasa.gov/archive/allData/61/MYD09/",
					"file_prefix":  "MYD09",
				},
			},
		},
	}
)
