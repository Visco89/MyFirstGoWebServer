package repository

import (
	"MyFirstGoWebServer/internal/core"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type jobOfferRepositoryImpl struct {
	db         *mongo.Database
	collection *mongo.Collection
}

func NewJobOfferRepository(db *mongo.Database) core.JobOfferRepository {
	return &jobOfferRepositoryImpl{
		db:         db,
		collection: db.Collection("jobOffers"),
	}
}

func (j jobOfferRepositoryImpl) GetByID(ctx context.Context, id string) (core.JobOffer, error) {
	var jobOffer core.JobOffer
	filter := bson.M{"_id": id}
	documentReturned := j.collection.FindOne(ctx, filter)
	err := documentReturned.Decode(&jobOffer)
	log.Println("Get job offer document", jobOffer, err, filter)
	return jobOffer, err
}

func (j jobOfferRepositoryImpl) Store(ctx context.Context, p *core.JobOffer) error {
	filter := bson.M{"_id": p.Id}
	if result := j.collection.FindOneAndReplace(ctx, filter, p); result != nil {
		_, err := j.collection.InsertOne(ctx, p)
		if err != nil {
			log.Println("Error on inserting new job offer", err)
			return err
		}
	}
	return nil
}

func (j jobOfferRepositoryImpl) Fetch(ctx context.Context, roleFilter string, companyFilter string) ([]*core.JobOffer, error) {
	var results []*core.JobOffer
	filter := bson.M{}
	if roleFilter != "" {
		filter["role"] = roleFilter
	}
	if companyFilter != "" {
		filter["company"] = companyFilter
	}

	cur, err := j.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var elem core.JobOffer
		err := cur.Decode(&elem)
		if err != nil {
			log.Println("Error while decoding jobOffer", err)
		}
		results = append(results, &elem)
	}
	return results, err
}

func (j jobOfferRepositoryImpl) PushApplication(ctx mongo.SessionContext, id string, applicationId string) error {
	result, err := j.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$addToSet": bson.M{"applications": applicationId}})
	if err != nil {
		log.Println("cannot update job offer", err)
		return err
	}
	if result.MatchedCount != 1 || result.ModifiedCount != 1 {
		return errors.New("no job offer modified")
	}
	return nil
}

func (j jobOfferRepositoryImpl) PopApplication(ctx mongo.SessionContext, applicationId string) error {
	result, err := j.collection.UpdateMany(ctx, bson.M{}, bson.M{"$pull": bson.M{"applications": applicationId}})
	if err != nil {
		log.Println("cannot update job offer", err)
		return err
	}
	if result.MatchedCount < 1 || result.ModifiedCount < 1 {
		return errors.New("no job offer modified")
	}
	return nil
}
