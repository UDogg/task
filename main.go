package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Completed bool               `json:"completed"`
	Title     string             `json:"title"`
	Body      string             `json:"body"`
}

var collection *mongo.Collection

func main() {
	fmt.Println("hello world")

	if os.Getenv("ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file:", err)
		}
	}

	MONGODB_URI := os.Getenv("MONGODB_URI")
	fmt.Println("MONGODB_URI:", MONGODB_URI)

	clientOptions := options.Client().ApplyURI(MONGODB_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MONGODB ATLAS")

	collection = client.Database("golang_db").Collection("todos")

	app := fiber.New()

	// Default CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	app.Patch("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)
	app.Get("/api/todos/:id", getTodoByID)
	app.Put("/api/todos/:id", replaceTodo)

	port := os.Getenv("PORT")
	fmt.Println("PORT:", port)
	if port == "" {
		port = "5001"
	}

	if os.Getenv("ENV") == "production" {
		app.Static("/", "./client/dist")
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))
}

func getTodos(c *fiber.Ctx) error {
	var todos []Todo

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Println("Find error:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to find todos"})
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var todo Todo
		if err := cursor.Decode(&todo); err != nil {
			log.Println("Cursor Decode error:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to decode todo"})
		}
		todos = append(todos, todo)
	}

	if err := cursor.Err(); err != nil {
		log.Println("Cursor iteration error:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Cursor iteration error"})
	}

	return c.JSON(todos)
}

func createTodo(c *fiber.Ctx) error {
	todo := new(Todo)

	if err := c.BodyParser(todo); err != nil {
		log.Println("BodyParser error:", err)
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if todo.Title == "" {
		log.Println("Empty Title error")
		return c.Status(400).JSON(fiber.Map{"error": "Todo title cannot be empty"})
	}

	if todo.Body == "" {
		log.Println("Empty Body error")
		return c.Status(400).JSON(fiber.Map{"error": "Todo body cannot be empty"})
	}

	if todo.ID.IsZero() {
		todo.ID = primitive.NewObjectID()
	}

	insertResult, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		log.Println("InsertOne error:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create todo"})
	}

	log.Println("Todo inserted:", insertResult.InsertedID)
	return c.Status(201).JSON(todo)
}

func updateTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		log.Println("Invalid todo ID error:", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
	}

	var updates bson.M
	if err := c.BodyParser(&updates); err != nil {
		log.Println("BodyParser error:", err)
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": updates}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Println("UpdateOne error:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update todo"})
	}

	return c.Status(200).JSON(fiber.Map{"success": true})
}

func deleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		log.Println("Invalid todo ID error:", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
	}

	filter := bson.M{"_id": objectID}
	deleteResult, err := collection.DeleteOne(context.Background(), filter)

	if err != nil {
		log.Println("DeleteOne error:", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete todo"})
	}

	if deleteResult.DeletedCount == 0 {
		log.Println("Todo not found:", id)
		return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
	}

	log.Println("Todo deleted:", id)
	return c.Status(200).JSON(fiber.Map{"success": true})
}

func getTodoByID(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
	}

	var todo Todo
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&todo)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
	}

	return c.JSON(todo)
}

func replaceTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
	}

	var todo Todo
	if err := c.BodyParser(&todo); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	todo.ID = objectID
	filter := bson.M{"_id": objectID}
	_, err = collection.ReplaceOne(context.Background(), filter, todo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to replace todo"})
	}

	return c.JSON(todo)
}
