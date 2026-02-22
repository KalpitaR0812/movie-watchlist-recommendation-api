package repositories

import (
	"context"
	"movie-watchlist/internal/database"
	"movie-watchlist/internal/models"
	

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	db *database.MongoDB
}

func NewUserRepository(db *database.MongoDB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	ctx := context.Background()
	collection := r.db.GetCollection("users")
	
	user.CreatedAt = getCurrentTime()
	user.UpdatedAt = getCurrentTime()
	
	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	
	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("users")
	
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id primitive.ObjectID) (*models.User, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("users")
	
	var user models.User
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("users")
	
	var user models.User
	err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
