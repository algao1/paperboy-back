package guardian

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"paperboy-back"
	"regexp"
	"time"

	"github.com/algao1/basically/btrank"
	"github.com/algao1/basically/document"
	"github.com/algao1/basically/parser"
	"github.com/algao1/basically/trank"
)

// Service represents an implementation of paperboy.GuardianService.
type Service struct {
	Key string
}

var _ paperboy.GuardianService = (*Service)(nil)

// Fetch returns the result of querying the Guardian API with the specified parameters.
func (s *Service) Fetch(qparams map[string]string) (*paperboy.Guardian, error) {
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

	var g paperboy.Guardian
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
	// Find image and add the appropriate caption.
	var im paperboy.Image
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

	// Get authors.
	authors := make([]string, 0)
	for _, tag := range r.Tags {
		authors = append(authors, tag.Title)
	}

	// Eliminate HTML tags from TrailText.
	reg := regexp.MustCompile("<[^>]*>")
	tText := reg.ReplaceAllString(r.Fields.TrailText, "")

	// Set up basically document.
	doc, err := document.Create(r.Fields.BodyText, &btrank.BiasedTextRank{}, &trank.KWTextRank{}, &parser.Parser{})
	if err != nil {
		return nil, fmt.Errorf("%q: %w", "unable to create basically document", err)
	}

	// Summarization using basically.
	sents, err := doc.Summarize(7, 0, "")
	if err != nil {
		return nil, fmt.Errorf("%q: %w", "unable to summarize document", err)
	}
	psents := make([]*paperboy.Sentence, len(sents))
	for idx, sen := range sents {
		psents[idx] = &paperboy.Sentence{Sentence: sen.Raw, Sentiment: sen.Sentiment}
	}

	// Keyword extraction using basically.
	kwords, err := doc.Highlight(10, true)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", "unable to extract keywords", err)
	}
	pkwords := make([]*paperboy.Keyword, len(kwords))
	for idx, w := range kwords {
		pkwords[idx] = &paperboy.Keyword{Word: w.Word, Weight: w.Weight}
	}

	fLength, sLength := doc.Characters()

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
			Title:       r.Title,
			TrailText:   tText,
			SummaryText: psents,
			Keywords:    pkwords,
			FullLength:  fLength,
			SummLength:  sLength,
		},
		Image: im,
	}
	log.Printf("[%s] summarization complete\n", r.Title)

	return &summ, nil
}
