package audit

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service interface {
	List(ctx context.Context) ([]Response, error)
}

type auditService struct {
	auditCol *mongo.Collection
}

func NewService(auditCol *mongo.Collection) Service {
	return &auditService{auditCol: auditCol}
}

func (s *auditService) List(ctx context.Context) ([]Response, error) {
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "timestamp", Value: -1}})
	findOptions.SetLimit(100)

	cursor, err := s.auditCol.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []Response = []Response{}
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}
