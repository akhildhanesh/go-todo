package middleware

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func JwtMiddleware(c *fiber.Ctx) error {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))

    tokenString := c.Get("Authorization")

    if tokenString == "" {
        return c.Status(401).JSON(fiber.Map{"error": "Missing or invalid token"})
    }


    tokenString = tokenString[len("Bearer "):]


    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return secretKey, nil
    })

    if err != nil || !token.Valid {
        return c.Status(401).JSON(fiber.Map{"error": "Invalid or expired token"})
    }

	id, err := token.Claims.GetSubject()

	if err != nil {
		fmt.Println("token subject error")
		return c.Status(401).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	c.Locals("userID", id)

    return c.Next()
}
