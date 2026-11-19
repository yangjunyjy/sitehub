package contract

// ProductDownload is resolved from the current published manifest, never a client URL.
type ProductDownload struct {
	URL      string `json:"url"`
	FileName string `json:"fileName"`
}
