package req

import (
	"UrlShortenerBackend/pkg/res"
	"net/http"
)

func HandleBody[T any](w *http.ResponseWriter, req *http.Request) (*T, error) {

	body, err := Decode[T](req.Body)
	defer req.Body.Close()
	if err != nil {
		res.Json(*w, err, http.StatusBadRequest)
		return nil, err
	}
	err = IsValid(body)
	if err != nil {
		res.Json(*w, err, http.StatusBadRequest)
		return nil, err
	}
	return &body, nil
}
