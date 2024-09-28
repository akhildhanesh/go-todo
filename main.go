package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

type Todo struct {
	ID int `json:"id"`
	Title string `json:"title"`
	Done bool `json:"done"`
	Body string `json:"body"`
}

func main() {
	app := fiber.New()

	todos := []Todo{}

	app.Get("/healthCheck", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Post("/api/todos", func(c *fiber.Ctx) error {
		todo := &Todo{}

		if err := c.BodyParser(todo); err != nil {
			return err
		}

		todo.ID = len(todos) + 1

		todos = append(todos, *todo)

		return c.JSON(todos)
	})

	app.Patch("/api/todos/:id/done", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")

		if err != nil {
			return c.Status(400).SendString("Invalid ID")
		}

		// todos[id - 1].Done = true

		for i, t := range todos {
			if t.ID == id {
				todos[i].Done = true
				break
			}
		}

		return c.JSON(todos)
	})

	app.Get("/api/todos", func(c *fiber.Ctx) error {
		return c.JSON(todos)
	})

	port := os.Getenv("PORT")

	if port == "" {
		port = ":3000"
	}

	log.Fatal(app.Listen(port))
}