package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// Config represents Cloudflare configuration for a specific domain.
type Config struct {
	Name   string
	Key    string
	Email  string
	ZoneID string
	URL    string
}

// Cloudflare is a partial implementation of the Cloudflare API.
type Cloudflare struct {
	Domains map[string]*Config
}

// Response from API
type Response struct {
	Success  bool                `json:"success"`
	Errors   []string            `json:"errors"`
	Messages []string            `json:"messages"`
	Result   []map[string]string `json:"result"`
}

// PurgeFile purges the given path on all supplied domains.
func (c Cloudflare) PurgeFile(path string) error {
	client := &http.Client{}
	base := "https://api.cloudflare.com/client/v4/zones/%s/purge_cache"
	for k := range c.Domains {
		config := c.Domains[k]
		url := strings.Trim(config.URL, "/") + "/" + strings.Trim(path, "/")
		files := map[string][]string{"files": []string{url}}
		b, err := json.Marshal(files)
		if err != nil {
			log.Println("JSON encoding error", err)
			return err
		}

		req, err := http.NewRequest("DELETE", fmt.Sprintf(base, config.ZoneID), bytes.NewReader(b))

		req.Header.Set("X-Auth-Email", config.Email)
		req.Header.Set("X-Auth-Key", config.Key)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		defer resp.Body.Close()
		if err != nil {
			return err
		}

		// decode response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		// check status code
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Cloudflare purge failed with %d - response: %s", resp.StatusCode, string(body))
		}

		// verify response
		var v Response
		json.Unmarshal(body, &v)
		if !v.Success {
			return fmt.Errorf("Cloudflare purge failed. Dumping response: %s", string(body))
		}

		log.Println("Cloudflare: purge path", url)
	}

	return nil
}
