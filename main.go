package main

import (
	"context"
	"gin-scalable-api/handlers"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
)

// We use the bson tag to map a struct field to the document attribute in the MongoDB collection

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
	collection := client.Database("recipes_db").Collection("recipes")
	recipesHandler = handlers.NewRecipesHandler(ctx, collection)
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

var recipesHandler *handlers.RecipesHandler

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
