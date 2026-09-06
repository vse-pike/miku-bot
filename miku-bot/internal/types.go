package internal

type DownloadRequest struct {
	URL string `json:"url"`
}

type DownloadResponse struct {
	Result      bool   `json:"result"`
	Description string `json:"description"`
	Path        string `json:"path"`
}
