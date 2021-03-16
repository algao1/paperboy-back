package guardian

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"paperboy-back"
	"regexp"
	"strconv"
	"time"

	"github.com/algao1/basically"
)

// Service represents an implementation of paperboy.GuardianService.
type Service struct {
	Key string
}

// Fetch returns the result of querying the Guardian API with the specified parameters.
func (s *Service) Fetch(qparams map[string]string) (*paperboy.Guardian, error) {
	var g paperboy.Guardian

	// Appends params onto url.
	url := "https://content.guardianapis.com/search?api-key=" + s.Key
	for k, v := range qparams {
		url += fmt.Sprintf("&%s=%s", k, v)
	}

	// Sends a GET request to url.
	res, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", "getting Guardian API failed", err)
	}
	log.Println("[Guardian API] response found")

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", "failed to read response body", err)
	}
	defer res.Body.Close()
	log.Println("[Guardian API] reading body")

	err = json.Unmarshal(body, &g)
	if err != nil {
		return nil, err
	}
	log.Println("[Guardian API] json unmarshaled")

	// Checks if response was OK.
	if g.Response.Status != "ok" {
		return nil, fmt.Errorf(g.Response.Status)
	}
	log.Println("[Guardian API] response ok")

	return &g, nil
}

// Extract returns a slice of summaries corresponding with the given slice of paperboy.Result.
func (s *Service) Extract(res []*paperboy.Result) ([]*paperboy.Summary, error) {
	var summaries []*paperboy.Summary
	for _, r := range res {
		summ, err := s.ExtractOne(r)
		if err != nil {
			log.Printf("%v", err)
			continue
		}
		summaries = append(summaries, summ)
	}
	return summaries, nil
}

// ExtractOne returns the result of summarizing a paperboy.Result.
func (s *Service) ExtractOne(r *paperboy.Result) (*paperboy.Summary, error) {
	var im paperboy.Image

	// Find image and add the appropriate caption.
	var assets []paperboy.Asset
	if len(r.Blocks.Main.Elements) > 0 && r.Blocks.Main.Elements[0].Type == "image" {
		assets = r.Blocks.Main.Elements[0].Assets
		im = paperboy.Image{Caption: r.Blocks.Main.Elements[0].ImgData.Caption}
	}
	for _, a := range assets {
		if a.TypeData.Width == 1000 {
			log.Printf("[%s] image found\n", r.Title)
			im.ImageFileURL = a.File
			break
		}
	}

	// Find and convert date string to time.Time.
	const layout = "2006-01-02T15:04:05Z"
	date, err := time.Parse(layout, r.Date)
	if err != nil {
		return nil, fmt.Errorf("[%s] could not parse date", r.Title)
	}

	// Convert full wordcount to int.
	fWordCount, err := strconv.Atoi(r.Fields.WordCount)
	if err != nil {
		return nil, fmt.Errorf("[%s] could not convert wordcount", r.Title)
	}

	// Get authors.
	authors := make([]string, 0)
	for _, tag := range r.Tags {
		authors = append(authors, tag.Title)
	}

	// Eliminate HTML tags from TrailText.
	reg := regexp.MustCompile("<[^>]*>")
	tText := reg.ReplaceAllString(r.Fields.TrailText, "")

	// Perform summarization using basically.
	nodes, err := basically.Summarize(r.Fields.BodyText, 7, basically.WithoutMergeQuotations())
	if err != nil {
		return nil, fmt.Errorf("%q: %w", "unable to complete summarization", err)
	}

	sents := make([]string, len(nodes))
	for idx, node := range nodes {
		sents[idx] = node.Sentence
		// log.Printf("[%.2f] %s\n", node.Score, node.Sentence)
	}

	summ := paperboy.Summary{
		Info: paperboy.Info{
			ContentID:   r.ContentID,
			SectionID:   r.SectionID,
			SectionName: r.SectionName,
			URL:         r.URL,
			Authors:     authors,
			Date:        date,
		},
		Article: paperboy.Article{
			Title:         r.Title,
			TrailText:     tText,
			SummaryText:   sents,
			FullWordCount: fWordCount,
			SummWordCount: 0,
		},
		Image: im,
	}
	log.Printf("[%s] summarization complete\n", r.Title)

	return &summ, nil
}
