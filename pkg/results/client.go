package results

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
)

const apiBaseURL = "https://37latuqm766patrerdf5rvdhqe0wgrug.lambda-url.us-east-1.on.aws/?"

var headers = http.Header{
	"Accept":          []string{"application/json, text/plain, */*"},
	"Accept-Language": []string{"es,en-US;q=0.9,en;q=0.8,pt;q=0.7"},
	"Connection":      []string{"keep-alive"},
	"Content-Type":    []string{"application/json"},
	"Origin":          []string{"https://resultadospresidencialesvenezuela2024.com"},
	"Referer":         []string{"https://resultadospresidencialesvenezuela2024.com/"},
	"Sec-Fetch-Dest":  []string{"empty"},
	"Sec-Fetch-Mode":  []string{"cors"},
	"Sec-Fetch-Site":  []string{"cross-site"},
	"User-Agent": []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"},
	"sec-ch-ua":          []string{"\"Not/A)Brand\";v=\"8\", \"Chromium\";v=\"126\", \"Google Chrome\";v=\"126\""},
	"sec-ch-ua-mobile":   []string{"?0"},
	"sec-ch-ua-platform": []string{"\"macOS\""},
}

func NewResultsRepo(store ResultStore, recaptcha string) ResultRepo {
	return &resultsV1Repo{
		storage:   store,
		recaptcha: recaptcha,
	}
}

type resultsV1Repo struct {
	storage   ResultStore
	recaptcha string
}

func (r *resultsV1Repo) FindByDocID(docID int) (*FindByDocIDResponse, error) {
	// Find the result by document id (VE Cedula)
	resultURL, err := findResultURLByDocID(docID, r.recaptcha)
	if err != nil {
		return nil, fmt.Errorf("error finding result URL by doc ID: %w", err)
	}

	// Find in cache (To avoid downloading the same result multiple times)
	stored, err := r.storage.GetResultByURL(resultURL)
	if err != nil && !errors.Is(err, ErrResultNotFound) {
		return nil, fmt.Errorf("error getting result by URL: %w", err)
	}

	// If the result is already stored, return it
	if stored != nil {
		isUpdated := false
		if stored.ExampleDocID == "" {
			isUpdated = true
			stored.ExampleDocID = strconv.Itoa(docID)
			err = r.storage.StoreResult(*stored)
			if err != nil {
				return nil, fmt.Errorf("error storing result: %w", err)
			}
		}
		return &FindByDocIDResponse{
			Result:    stored,
			IsCached:  true,
			IsUpdated: isUpdated,
		}, nil
	}

	// Download the result if it's not stored
	downloadedResultBytes, err := r.downloadResultURLByURL(resultURL)
	if err != nil {
		return nil, fmt.Errorf("error downloading result: %w", err)
	}

	downloadedResult := Result{
		Url:          resultURL,
		Bytes:        downloadedResultBytes,
		ExampleDocID: strconv.Itoa(docID),
	}

	// Store the downloaded result in storage
	err = r.storage.StoreResult(downloadedResult)
	if err != nil {
		return nil, fmt.Errorf("error storing result: %w", err)
	}
	return &FindByDocIDResponse{
		Result:    &downloadedResult,
		IsCached:  false,
		IsUpdated: false,
	}, nil
}

func (*resultsV1Repo) downloadResultURLByURL(url string) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodGet)

	// Add headers
	for k, v := range headers {
		req.Header.Add(k, v[0])
	}

	err := fasthttp.Do(req, res)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	return res.Body(), nil
}

func findResultURLByDocID(docID int, recaptcha string) (string, error) {
	url := apiBaseURL + "cedula=V" + strconv.Itoa(docID) + "&recaptcha=" + recaptcha

	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()

	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodGet)

	// Add headers
	for k, v := range headers {
		req.Header.Add(k, v[0])
	}

	err := fasthttp.Do(req, res)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}

	if res.StatusCode() == fasthttp.StatusBadGateway {
		return "", ErrResultNotFound
	}
	body := res.Body()
	if res.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("unexpected response: %s", string(body))
	}

	var resultMap map[string]string
	err = json.Unmarshal(body, &resultMap)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling response body: %w", err)
	}
	return resultMap["url"], nil
}
