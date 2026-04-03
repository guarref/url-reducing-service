package web

type reduceURLRequest struct {
	URL string `json:"url"`
}

type reduceURLResponse struct {
	URL       string `json:"url"`
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

type getOriginalURLResponse struct {
	URL string `json:"url"`
}
