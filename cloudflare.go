package cloudflare

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
)

// CloudflareConfig represents a single cloudflare domain.
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

// PurgeFile purges the given path on all supplied domains.
func (c Cloudflare) PurgeFile(path string) error {
    client := &http.Client{}
    base := "https://api.cloudflare.com/client/v4/zones/%s/purge_cache"
    for k := range c.Domains {
        config := c.Domains[k]
        files := map[string][]string{"files": []string{config.URL + "/" + path}}
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

        body, err := ioutil.ReadAll(resp.Body)
        if resp.StatusCode != http.StatusOK {
            return fmt.Errorf("Cloudflare purge failed with %d - response: %s", resp.StatusCode, string(body))
        }

        log.Println("Cloudflare: purge path", path, "on domain", k)
    }

    return nil
}
