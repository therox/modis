package modis

import (
	"bytes"
	"io"
)

// def get_modis_data_with_token(self, download_url: str, download: bool = False) -> Optional[str]:
// """Downloads MODIS data from url with token

// Args:
// 	download_url(str): url to download
// 	download(bool): if True, download file from url (true) or return it's content (false)

// Returns:
// 	str: file content or None if download is False

// """
// with requests.Session() as session:
// 	session.headers["Authorization"] = f"Bearer {self._modis_token}"

// 	response = session.get(download_url)
// 	# extract the filename from the url to be used when saving the file
// 	filename = download_url[download_url.rfind("/") + 1 :]

// 	if response.ok:
// 		if download:
// 			# save the file
// 			with open(filename, "wb") as fd:
// 				for chunk in response.iter_content(chunk_size=1024 * 1024):
// 					fd.write(chunk)
// 		else:
// 			return response.text
// 	else:
// 		raise ValueError(
// 			f'Error downloading file "{download_url}". Status Code: {response.status_code}. Content:'
// 			f" {response.text}"
// 		)
// return None

func (c *Client) GetModisScene(GranuleID string) (io.Reader, error) {
	res := bytes.NewReader([]byte(""))
	return res, nil
}
