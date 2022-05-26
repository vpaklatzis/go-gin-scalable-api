package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
	"os"
	"time"
)

// We use the bson tag to map a struct field to the document attribute in the MongoDB collection
type Recipe struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PublishedAt  time.Time          `json:"publishedAt" bson:"publishedAt"`
}

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
	var recipe Recipe
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
	recipes := make([]Recipe, 0)
	for cursor.Next(handler.ctx) {
		var recipe Recipe
		cursor.Decode(&recipe)
		recipes = append(recipes, recipe)
	}
	fmt.Printf("ClientIP: %s\n", context.ClientIP())
	context.JSON(http.StatusOK, recipes)
}

func (handler *RecipesHandler) UpdateRecipesHandler(context *gin.Context) {
	id := context.Param("id")
	var recipe Recipe
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
	var recipe Recipe
	err := cursor.Decode(&recipe)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, recipe)
}

// Initializes the recipes variable. Reads the recipes.json file and converts the content
// into an array of recipes. Gets executed at the application startup.
// Connects to mongodb. Reads recipes.json and stores the recipes in the recipes collection.
func init() {
	//recipes = make([]Recipe, 0)
	//file, _ := ioutil.ReadFile("recipes.json")
	//_ = json.Unmarshal([]byte(file), &recipes)
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal("Error occurred while trying to connect to mongodb: ", err)
	}
	log.Println("Connected to mongodb!")
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	recipesHandler = NewRecipesHandler(ctx, collection)
	//var listOfRecipes []interface{}
	//for _, recipe := range recipes {
	//	listOfRecipes = append(listOfRecipes, recipe)
	//}
	//collection := client.Database("recipes_db").Collection("recipes")
	//InsertManyResult, err := collection.InsertMany(ctx, listOfRecipes)
	//if err != nil {
	//log.Fatal("Could not insert recipes to db: ", err)
	//}
	//log.Println("Inserted recipes: ", len(InsertManyResult.InsertedIDs))
}

var recipesHandler *RecipesHandler

func main() {
	router := gin.Default()
	router.SetTrustedProxies([]string{"localhost"})
	router.POST("/recipes", recipesHandler.CreateRecipesHandler)
	router.GET("/recipes", recipesHandler.ListRecipesHandler)
	router.PUT("/recipes/:id", recipesHandler.UpdateRecipesHandler)
	router.DELETE("/recipes/:id", recipesHandler.DeleteRecipesHandler)
	router.GET("recipes/:id", recipesHandler.GetOneRecipesHandler)
	router.Run()
}
