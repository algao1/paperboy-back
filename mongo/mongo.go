package mongo

import (
	"context"
	"fmt"
	"paperboy-back"

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
func Open(uri, key, db, col string) (paperboy.SummaryService, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(fmt.Sprintf(uri, key)))
	if err != nil {
		return nil, fmt.Errorf("%q: %w", "could not connect to collection", err)
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
// 		startID: nil -> most recent article
func (s *SummaryService) Summaries(sectionID, startID string, size int) ([]*paperboy.Summary, string, error) {
	// Query range filter using the default indexed (objectid) _id field and sectionid.
	var objectID primitive.ObjectID
	var err error

	filters := bson.M{}
	if len(startID) > 0 {
		objectID, err = primitive.ObjectIDFromHex(startID)
		if err != nil {
			return nil, "", fmt.Errorf("%q: %w", "invalid objectId", err)
		}
		filters["_id"] = bson.M{"$lt": objectID}
	}
	if len(sectionID) > 0 {
		filters["info.sectionid"] = sectionID
	}

	// Query options.
	var opts []*options.FindOptions
	opts = append(opts, options.Find().SetSort(bson.M{"_id": -1}))
	opts = append(opts, options.Find().SetLimit(int64(size)))

	// Fetch cursor.
	cursor, err := s.col.Find(context.TODO(), filters, opts...)
	if err != nil {
		return nil, "", fmt.Errorf("%q: %w", "cursor not found", err)
	}

	var res []*paperboy.Summary
	var lastValue string
	for cursor.Next(context.Background()) {
		var summ paperboy.Summary
		err = cursor.Decode(&summ)
		if err != nil {
			return res, lastValue, fmt.Errorf("%q: %w", "could not decode summary", err)
		}
		// Gets the objectId.
		var h hex
		cursor.Decode(&h)
		lastValue = h.ID.Hex()
		// Appends and updates the last objectId.
		summ.ObjectID = lastValue
		res = append(res, &summ)
	}

	return res, lastValue, nil
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
		return fmt.Errorf("%q: %w", "could not update document", err)
	}
	return nil
}
