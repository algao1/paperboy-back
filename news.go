package paperboy

// TypeData contains information about an asset.
type TypeData struct {
	Width  int
	Height int
}

// Asset is an image, video or resource used in an article.
type Asset struct {
	File     string
	TypeData TypeData
}

// ImageTypeData contains information about an image.
type ImageTypeData struct {
	Caption string
}

// Element contains a collection of assets, and associated information.
type Element struct {
	Type    string
	Assets  []Asset
	ImgData ImageTypeData `json:"imageTypeData"`
}

// Block contains a collection of elements.
type Block struct {
	Elements []Element
}

// Blocks contains blocks of content.
type Blocks struct {
	Main Block
}

// Fields are the metadata associated with the article's content.
type Fields struct {
	TrailText string
	BodyText  string
	WordCount string `json:"wordcount"`
}

// Tag is an associated metadata tag within an article.
type Tag struct {
	ID    string
	Type  string
	Title string `json:"webTitle"`
}

// Result contains information on individual articles.
type Result struct {
	ContentID   string `json:"id"`
	SectionID   string
	SectionName string
	URL         string `json:"webUrl"`
	Date        string `json:"webPublicationDate"`
	Title       string `json:"webTitle"`
	Fields      Fields
	Tags        []Tag
	Blocks      Blocks
}

// Response contains the HTTP response status, and a collection of results.
type Response struct {
	Status  string
	Results []*Result
}

// Guardian is the HTTP response from the Guardian API.
type Guardian struct {
	Response Response
}

// GuardianService defines the functionality provided by the service.
type GuardianService interface {
	Fetch(qparams map[string]string) (*Guardian, error)
	Extract(res []*Result) ([]*Summary, error)
	ExtractOne(r *Result) (*Summary, error)
}
