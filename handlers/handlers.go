package handlers

import (
	"context"
	"fmt"
	"gin-scalable-api/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

type RecipesHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection) *RecipesHandler {
	return &RecipesHandler{
		collection: collection,
		ctx:        ctx,
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
	context.JSON(http.StatusOK, recipe)
}

// ListRecipeHandler Find() fetches all the requested items from the recipe collection.
// Find() returns a cursor, or rather a stream of documents.
func (handler *RecipesHandler) ListRecipesHandler(context *gin.Context) {
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
