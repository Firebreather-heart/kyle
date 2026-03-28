package scraper

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type SearchResult struct {
	Title   string
	URL     string
	Snippet string
	Content string
}

func ExecuteSearch(query string) (string, error) {
	safeQuery := strings.ReplaceAll(query, " ", "+")
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", safeQuery)

	client := &http.Client{Timeout: 20 * time.Second}
	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("search network error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("duckduckgo rejected request: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	var results []SearchResult

	doc.Find(".result__body").Each(func(i int, s *goquery.Selection) {
		if i >= 3 {
			return
		}
		
		title := strings.TrimSpace(s.Find(".result__a").Text())
		snippet := strings.TrimSpace(s.Find(".result__snippet").Text())
		rawURL, exists := s.Find(".result__a").Attr("href")
		
		if exists && !strings.Contains(rawURL, "duckduckgo.com") {
			results = append(results, SearchResult{
				Title:   title,
				Snippet: snippet,
				URL:     rawURL,
			})
		}
	})

	if len(results) == 0 {
		return "No results found.", nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Research Results for '%s':\n\n", query))

	for i, res := range results {
		deepContent := deepReadURL(client, res.URL)
		
		if deepContent != "" {
			output.WriteString(fmt.Sprintf("--- SOURCE %d (Full Article) ---\nTitle: %s\nURL: %s\nContent:\n%s\n\n", i+1, res.Title, res.URL, deepContent))
		} else {
			output.WriteString(fmt.Sprintf("--- SOURCE %d (Snippet Only) ---\nTitle: %s\nURL: %s\nContent:\n%s\n\n", i+1, res.Title, res.URL, res.Snippet))
		}
	}

	return output.String(), nil
}

func deepReadURL(client *http.Client, targetURL string) string {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client.Timeout = 15 * time.Second 
	resp, err := client.Do(req)
	
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return ""
	}

	var articleText strings.Builder
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if len(text) > 40 {
			articleText.WriteString(text + " ")
		}
	})

	fullText := articleText.String()
	
	if len(fullText) > 5000 {
		return fullText[:5000] + "... [TRUNCATED]"
	} else {
		return fullText
	}
	
	return fullText
}