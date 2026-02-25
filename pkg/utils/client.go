package utils

import (
	"log/slog"
	"net/http"
	"time"
)


func Client() *http.Client {

	transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
}
    client := &http.Client{
		Transport: transport,
		Timeout: 20 * time.Second,
	 }
     return client
}

func GetRequest(url string) (*http.Response, error) {

   client := Client()
   req, err := http.NewRequest(http.MethodGet, url, nil)
   if err != nil {
	slog.Error("Error creating request", "error", err)
	return nil, err
   }

   //run request
   resp, err := client.Do(req)
   if err != nil {
	slog.Error("Error executing request", "error", err)
	return nil, err
   }
   defer resp.Body.Close()
   return resp, nil
}