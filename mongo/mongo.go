package mongo

import (
	"context"
	"fmt"
	"paperboy-back"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SummaryService is a MongoDB implementation of paperboy.SummaryService.
type SummaryService struct {
	col *mongo.Collection
}

var _ paperboy.SummaryService = (*SummaryService)(nil)

// Open returns a pointer to SummaryService with the MongoDB collection configured.
func Open(uri, key, db, col string) (*SummaryService, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(fmt.Sprintf(uri, key)))
	if err != nil {
		return nil, fmt.Errorf("%q: %w", "unable to connect to collection", err)
	}
	collection := client.Database(db).Collection(col)
	return &SummaryService{col: collection}, nil
}

// Summary returns a pointer to a summary object for a given objectID.
func (s *SummaryService) Summary(id string) (*paperboy.Summary, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", "invalid objectId", err)
	}

	var res paperboy.Summary
	err = s.col.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&res)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", "objectId not found", err)
	}
	return &res, nil
}

type hex struct {
	ID primitive.ObjectID `json:"_id" bson:"_id"`
}

// Summaries returns a slice of the most recent summaries with a given sectionID such as 'world' or 'tech'.
// A starting objectID and page number must also be provided for pagination.
//		sectionID: nil -> articles in all sections
func (s *SummaryService) Summaries(sectionID string, endDate time.Time, size int) ([]*paperboy.Summary, string, error) {
	// Query range filter using the default indexed (objectid) _id field and sectionid.
	var err error

	// Filters.
	filters := bson.M{}
	filters["info.date"] = bson.M{"$lt": endDate.UTC()}
	if len(sectionID) > 0 {
		filters["info.sectionid"] = sectionID
	}

	// Query options.
	var opts []*options.FindOptions
	opts = append(opts, options.Find().SetSort(bson.M{"info.date": -1}))
	opts = append(opts, options.Find().SetLimit(int64(size)))

	// Fetch cursor.
	cursor, err := s.col.Find(context.TODO(), filters, opts...)
	if err != nil {
		return nil, "", fmt.Errorf("%q: %w", "cursor not found", err)
	}

	var res []*paperboy.Summary
	for cursor.Next(context.Background()) {
		var summ paperboy.Summary
		err = cursor.Decode(&summ)
		if err != nil {
			return res, "", fmt.Errorf("%q: %w", "unable to decode summary", err)
		}
		// Gets and updates the last objectId.
		var h hex
		cursor.Decode(&h)
		summ.ObjectID = h.ID.Hex()
		res = append(res, &summ)
	}

	retDate := time.Now().UTC()
	if len(res) > 0 {
		retDate = res[len(res)-1].Info.Date.UTC()
	}

	return res, retDate.Format(time.RFC3339), nil
}

// Search returns a list of summaries found using Mongo's fuzzy search
// with the text index being the 'keywords' generated previously.
func (s *SummaryService) Search(query string, size int) ([]*paperboy.Summary, error) {
	var err error

	// Configure search options, and filter.
	filters := bson.M{"$text": bson.M{"$search": query}}
	opts := []*options.FindOptions{
		options.Find().SetProjection(bson.M{"score": bson.M{"$meta": "textScore"}}),
		options.Find().SetSort(bson.M{"score": bson.M{"$meta": "textScore"}}),
		options.Find().SetLimit(int64(size)),
	}

	cursor, err := s.col.Find(context.TODO(), filters, opts...)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", "unable to perform search", err)
	}

	var summaries []*paperboy.Summary
	for cursor.Next(context.Background()) {
		var summ paperboy.Summary
		err = cursor.Decode(&summ)
		if err != nil {
			return nil, fmt.Errorf("%q: %w", "unable to decode summary", err)
		}
		// Gets and updates the last objectId.
		var h hex
		cursor.Decode(&h)
		summ.ObjectID = h.ID.Hex()
		summaries = append(summaries, &summ)
	}

	return summaries, nil
}

// Create inserts a summary into the database if possible, otherwise,
// it will update the existing entry.
func (s *SummaryService) Create(summary *paperboy.Summary) error {
	// Configure options, filter, and update.
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"info.contentid": summary.Info.ContentID}
	update := bson.M{"$set": summary}

	_, err := s.col.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return fmt.Errorf("%q: %w", "unable to update document", err)
	}
	return nil
}
