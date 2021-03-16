package paperboy

import "time"

// Info contains meta information about the article such as the contentId,
// sectionId, sectionName, url, authors, and date of publication.
type Info struct {
	ContentID   string `json:"ContentId"`
	SectionID   string `json:"SectionId"`
	SectionName string
	URL         string
	Authors     []string
	Date        time.Time
}

// Article contains information derived from the article such as the title,
// trail text, summary text, and wordcount.
type Article struct {
	Title         string
	TrailText     string
	SummaryText   []string
	FullWordCount int
	SummWordCount int
}

// Image contains information about the image used in the article.
type Image struct {
	ImageFileURL string
	Caption      string
}

// Summary contains all the relevant information including objectid, metadata,
// article, data, and image data.
type Summary struct {
	ObjectID string `json:"ObjectId" bson:"-"`
	Info     Info
	Article  Article
	Image    Image
}

// SummaryService defines the functionality provided by the service.
//	Summary: returns a summary with a given objectID.
//	Summaries: returns a list of summaries matching a sectionID and starting from
//		a startID. If startID is empty, will fetch the most recent entries.
// 	Create: writes a summary to the database.
type SummaryService interface {
	Summary(objectID string) (*Summary, error)
	Summaries(sectionID, startID string, size int) ([]*Summary, string, error)
	Create(s *Summary) error
}
