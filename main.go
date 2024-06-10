package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	Body      string `json:"body"`
}

func main() {
	app := fiber.New()

	todos := []Todo{}

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"msg": "Hello, World!"})
	})

	app.Post("/api/todos", func(c *fiber.Ctx) error {
		todo := Todo{}
		if err := c.BodyParser(&todo); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		if todo.Body == "" || todo.Title == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Title and Body are required"})
		}
		todo.ID = len(todos) + 1
		todos = append(todos, todo)
		return c.Status(201).JSON(todo)
	})
	// var todo Todo
	// if err := c.BodyParser(&todo); err != nil {
	// 	return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	// }
	// todo.ID = len(todos) + 1
	// todos = append(todos, todo)
	// return c.Status(201).JSON(todo)
	log.Fatal(app.Listen(":4000"))

}
