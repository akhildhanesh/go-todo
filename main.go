package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Todo struct {
	ID int `gorm:"primaryKey" json:"id"`
	Title string `gorm:"not null" json:"title"`
	Done bool `json:"done"`
	Body string `json:"body"`
}

type User struct {
    ID        uint   `gorm:"primaryKey"`
    FirstName string `gorm:"not null"`
    LastName  string `gorm:"not null"`
    Email     string `gorm:"unique;not null"`
}

func main() {
	dsn := "root:@tcp(127.0.0.1:3306)/go?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        fmt.Println("Failed to connect to the database:", err)
        return
    }

	err = db.AutoMigrate(&User{}, &Todo{})
    if err != nil {
        fmt.Println("Failed to migrate database:", err)
        return
    }

	app := fiber.New()

	todos := []Todo{}

	app.Get("/healthCheck", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Post("/api/createUser", func(c *fiber.Ctx) error {
		user := &User{}

		if err := c.BodyParser(user); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid data"})
		}

		if err := db.Create(user).Error; err != nil {
			fmt.Println("Failed to save user", err)
		} else {
			fmt.Println("saved", user)
		}
		
		return c.Status(201).JSON(fiber.Map{"success": "true", "message": "user created"})
	})

	app.Post("/api/todos", func(c *fiber.Ctx) error {
		todo := &Todo{}

		if err := c.BodyParser(todo); err != nil {
			return err
		}

		// todo.ID = len(todos) + 1

		// todos = append(todos, *todo)
		if err := db.Create(todo).Error; err != nil {
			fmt.Println("Failed to save Todo", err)
		} else {
			fmt.Println("saved", todo)
		}

		return c.JSON(todos)
	})

	app.Patch("/api/todos/:id/done", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")

		if err != nil {
			return c.Status(400).SendString("Invalid ID")
		}

		// todos[id - 1].Done = true

		// for i, t := range todos {
		// 	if t.ID == id {
		// 		todos[i].Done = true
		// 		break
		// 	}
		// }
		var todo Todo
		if err := db.First(&todo, id).Error; err != nil {
			fmt.Println("Failed to Get Todo", err)
			return c.Status(400).SendString("Invalid ID")
		}

		todo.Done = true

		if err := db.Save(&todo).Error; err != nil {
			fmt.Println("Failed to Save Todo", err)
		}

		return c.JSON(todo)
	})

	app.Get("/api/todos", func(c *fiber.Ctx) error {
		var todos []Todo
		if err := db.Find(&todos).Error; err != nil {
			fmt.Println("Failed to Get Todo", err)
		} else {
			fmt.Println("Retrieved", todos)
		}
		return c.JSON(todos)
	})

	port := os.Getenv("PORT")

	if port == "" {
		port = ":3000"
	}

	log.Fatal(app.Listen(port))
}