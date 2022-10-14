package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"gin-scalable-api/models"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"time"
)

type RecipesHandler struct {
	collection  *mongo.Collection
	ctx         context.Context
	redisClient *redis.Client
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		collection:  collection,
		ctx:         ctx,
		redisClient: redisClient,
	}
}

func (handler *RecipesHandler) CreateRecipesHandler(context *gin.Context) {
	var recipe models.Recipe
	if err := context.ShouldBindJSON(&recipe); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := handler.collection.InsertOne(handler.ctx, recipe)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while inserting a new recipe",
		})
		return
	}
	log.Println("Remove data from Redis")
	handler.redisClient.Del(context, "recipes")
	context.JSON(http.StatusOK, recipe)
}

// ListRecipeHandler Find() fetches all the requested items from the recipe collection.
// Find() returns a cursor, or rather a stream of documents.
func (handler *RecipesHandler) ListRecipesHandler(context *gin.Context) {
	val, err := handler.redisClient.Get(context, "recipes").Result()
	if err == redis.Nil {
		log.Printf("Request to MongoDB")
		cursor, err := handler.collection.Find(handler.ctx, bson.M{})
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer cursor.Close(handler.ctx)
		recipes := make([]models.Recipe, 0)
		for cursor.Next(handler.ctx) {
			var recipe models.Recipe
			cursor.Decode(&recipe)
			recipes = append(recipes, recipe)
		}
		fmt.Printf("ClientIP: %s\n", context.ClientIP())
		context.JSON(http.StatusOK, recipes)
	} else if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		log.Printf("Request to Redis")
		recipes := make([]models.Recipe, 0)
		json.Unmarshal([]byte(val), &recipes)
		context.JSON(http.StatusOK, recipes)
	}
}

func (handler *RecipesHandler) UpdateRecipesHandler(context *gin.Context) {
	id := context.Param("id")
	var recipe models.Recipe
	if err := context.ShouldBindJSON(&recipe); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.UpdateOne(handler.ctx, bson.M{
		"_id": objectId,
	}, bson.D{{"$set", bson.D{
		{"name", recipe.Name},
		{"instructions", recipe.Instructions},
		{"ingredients", recipe.Ingredients},
		{"tags", recipe.Tags},
	}}})
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Recipe has been updated"})
}

func (handler *RecipesHandler) DeleteRecipesHandler(context *gin.Context) {
	id := context.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.DeleteOne(handler.ctx, bson.M{
		"_id": objectId,
	})
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "Recipe not found",
		})
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been deleted",
	})
}

func (handler *RecipesHandler) GetOneRecipesHandler(context *gin.Context) {
	id := context.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	cursor := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": objectId,
	})
	var recipe models.Recipe
	err := cursor.Decode(&recipe)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, recipe)
}
