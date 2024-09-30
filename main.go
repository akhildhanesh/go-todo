package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"todo/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
    ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
    FirstName string `gorm:"not null" json:"firstName"`
    LastName  string `gorm:"not null" json:"lastName"`
    Email     string `gorm:"unique;not null" json:"email"`
	Password string `gorm:"not null" json:"password"`
    Todos     []Todo `gorm:"foreignKey:UserID"` // One-to-many relationship
}

type Todo struct {
    ID      int    `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID  uint   `gorm:"not null"` // Foreign key to User
    Title   string `gorm:"not null" json:"title"`
    Done    bool   `json:"done"`
    Body    string `json:"body"`
}

type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

var secretKey []byte

func createToken(id string) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": id,                    // Subject (user identifier)
			"iss": "todo-app",                  // Issuer
			"aud": "user",           // Audience (user role)
			"exp": time.Now().Add(time.Hour).Unix(), // Expiration time
			"iat": time.Now().Unix(),                 // Issued at
		})

	fmt.Printf("Token claims added: %+v\n", claims)
	tokenString, err := claims.SignedString(secretKey)
    if err != nil {
        return "", err
    }
	return tokenString, nil
}


func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

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

		if user.Password == "" {
			return c.Status(500).JSON(fiber.Map{"error": "Please enter your password"})
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
		}

		user.Password = string(hashedPassword)

		if err := db.Create(user).Error; err != nil {
			fmt.Println("Failed to save user", err)
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
		} else {
			fmt.Println("saved", user)
		}
		
		return c.Status(201).JSON(fiber.Map{"success": "true", "message": "user created"})
	})

	app.Post("/api/login", func(c *fiber.Ctx) error {
		loginReq := &LoginRequest{}

		if err := c.BodyParser(loginReq); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "ERROR"})
		}

		var user User

		if err := db.Where("email = ?", loginReq.Email).Find(&user).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
		}

		fmt.Println(user)

		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password))
		if err != nil {
			fmt.Println("Invalid password")
			return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
		}

		fmt.Println("id:", strconv.Itoa(int(user.ID)))

		tokenString, err := createToken(strconv.Itoa(int(user.ID)))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Server ERROR"})
		}

		return c.JSON(fiber.Map{"success": "true", "message": "Logged In", "token": tokenString})
	})

	app.Post("/api/todos", middleware.JwtMiddleware, func(c *fiber.Ctx) error {
		todo := &Todo{}

		if err := c.BodyParser(todo); err != nil {
			return err
		}

		fmt.Println("user id: ", c.Locals("userID"))

		userID, err := strconv.Atoi(c.Locals("userID").(string))

		if err != nil {
			fmt.Println("Failed to save Todo", err)
		}
		
		todo.UserID = uint(userID)

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

	app.Get("/api/todos", middleware.JwtMiddleware, func(c *fiber.Ctx) error {
		var todos []Todo
		if err := db.Where("user_id = ?", c.Locals("userID")).Find(&todos).Error; err != nil {
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