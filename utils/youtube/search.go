package youtube

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type VideoResult struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Duration  string `json:"duration"`
	Thumbnail string `json:"thumbnail"`
}

type youtubeiRequest struct {
	Context struct {
		Client struct {
			ClientName    string `json:"clientName"`
			ClientVersion string `json:"clientVersion"`
			Hl            string `json:"hl"`
			Gl            string `json:"gl"`
		} `json:"client"`
	} `json:"context"`
	Query string `json:"query"`
}

func SearchVideos(query string) ([]VideoResult, error) {
	reqBody := youtubeiRequest{
		Query: query,
	}
	reqBody.Context.Client.ClientName = "WEB"
	reqBody.Context.Client.ClientVersion = "2.20240101.00.00"
	reqBody.Context.Client.Hl = "fr"
	reqBody.Context.Client.Gl = "FR"

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	req, err := http.NewRequest("POST", "https://www.youtube.com/youtubei/v1/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status code %d: %s", resp.StatusCode, string(body))
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	var results []VideoResult
	extractVideos(raw, &results)

	return results, nil
}

func extractVideos(node interface{}, results *[]VideoResult) {
	switch val := node.(type) {
	case map[string]interface{}:
		if vr, ok := val["videoRenderer"].(map[string]interface{}); ok {
			var vid VideoResult

			if id, ok := vr["videoId"].(string); ok {
				vid.ID = id
			}

			if titleObj, ok := vr["title"].(map[string]interface{}); ok {
				if runs, ok := titleObj["runs"].([]interface{}); ok && len(runs) > 0 {
					if runMap, ok := runs[0].(map[string]interface{}); ok {
						vid.Title, _ = runMap["text"].(string)
					}
				}
			}

			if ownerObj, ok := vr["ownerText"].(map[string]interface{}); ok {
				if runs, ok := ownerObj["runs"].([]interface{}); ok && len(runs) > 0 {
					if runMap, ok := runs[0].(map[string]interface{}); ok {
						vid.Author, _ = runMap["text"].(string)
					}
				}
			}

			if lenObj, ok := vr["lengthText"].(map[string]interface{}); ok {
				vid.Duration, _ = lenObj["simpleText"].(string)
			}

			if thumbObj, ok := vr["thumbnail"].(map[string]interface{}); ok {
				if thumbs, ok := thumbObj["thumbnails"].([]interface{}); ok && len(thumbs) > 0 {
					if lastThumb, ok := thumbs[len(thumbs)-1].(map[string]interface{}); ok {
						vid.Thumbnail, _ = lastThumb["url"].(string)
					}
				}
			}

			if vid.ID != "" && vid.Title != "" {
				*results = append(*results, vid)
			}
			return
		}

		for _, v := range val {
			extractVideos(v, results)
		}

	case []interface{}:
		for _, item := range val {
			extractVideos(item, results)
		}
	}
}
