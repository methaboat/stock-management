package list

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	col *mongo.Collection
}

func NewStorage(col *mongo.Collection) *Storage {
	return &Storage{col: col}
}

func (s *Storage) Execute(ctx context.Context) ([]Response, error) {
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}}).SetLimit(100)

	cursor, err := s.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	logs := []Response{}
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}
