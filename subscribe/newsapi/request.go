package newsapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func RequestArticles(categoryName string) (*ArticlesWrapper, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://newsapi.org/v2/top-headlines?country=jp&pageSize=3&category=%s", categoryName), nil)
	req.Header.Set("X-Api-Key", os.Getenv("API_KEY"))
	var client http.Client
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New("request failed")
	}
	body, _ := ioutil.ReadAll(res.Body)
	var articlesWrapper ArticlesWrapper
	if err := json.Unmarshal(body, &articlesWrapper); err != nil {
		return nil, err
	}
	return &articlesWrapper, nil
}
