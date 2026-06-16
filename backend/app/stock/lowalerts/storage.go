package lowalerts

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Storage struct {
	col *mongo.Collection
}

func NewStorage(col *mongo.Collection) *Storage {
	return &Storage{col: col}
}

func (s *Storage) Execute(ctx context.Context) ([]Response, error) {
	filter := bson.M{"$expr": bson.M{"$lte": bson.A{"$availableStock", "$reorderPoint"}}}

	cursor, err := s.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	list := []Response{}
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}
