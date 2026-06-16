package list

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

func (s *Storage) Execute(ctx context.Context) ([]Reconciliation, error) {
	cursor, err := s.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	list := []Reconciliation{}
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}
