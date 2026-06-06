package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var configFile = flag.String("f", "user-svc.yaml", "the config file")

type Config struct {
	Name string `yaml:"name"`
	Port int    `yaml:"port"`
}

// User represents a user in the system
type User struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	TenantID     uint   `json:"tenantId"`
	Role         string `json:"role"`
}

// In-memory user store (replace with DB later)
var (
	users = make(map[string]*User)
	mu    sync.RWMutex
	nextUserID uint = 1
)

// JWT claims
type Claims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	TenantID uint   `json:"tenantId"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

const jwtSecret = "agentmesh-secret-key-change-in-production"

// generateJWT creates a new JWT token
func generateJWT(user *User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		TenantID: user.TenantID,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// checkPassword verifies a password against a hash
func checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func main() {
	flag.Parse()

	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Login endpoint
	r.POST("/api/user/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
			return
		}

		mu.RLock()
		user, exists := users[req.Username]
		mu.RUnlock()

		if !exists || !checkPassword(req.Password, user.PasswordHash) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid username or password"})
			return
		}

		token, err := generateJWT(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": token,
			"user": gin.H{
				"id":        user.ID,
				"username":  user.Username,
				"email":     user.Email,
				"tenantId":  user.TenantID,
				"role":      user.Role,
			},
		})
	})

	// Register endpoint
	r.POST("/api/user/register", func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required,min=6"`
			Email    string `json:"email" binding:"required,email"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request: " + err.Error()})
			return
		}

		mu.RLock()
		_, exists := users[req.Username]
		mu.RUnlock()

		if exists {
			c.JSON(http.StatusConflict, gin.H{"message": "Username already exists"})
			return
		}

		hash, err := hashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash password"})
			return
		}

		mu.Lock()
		user := &User{
			ID:           nextUserID,
			Username:     req.Username,
			Email:        req.Email,
			PasswordHash: hash,
			TenantID:     1, // Default tenant
			Role:         "user",
		}
		users[req.Username] = user
		nextUserID++
		mu.Unlock()

		c.JSON(http.StatusCreated, gin.H{
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
		})
	})

	// Get users (protected for demo)
	r.GET("/api/users", func(c *gin.Context) {
		mu.RLock()
		result := make([]*User, 0, len(users))
		for _, u := range users {
			result = append(result, u)
		}
		mu.RUnlock()
		c.JSON(http.StatusOK, gin.H{"users": result})
	})

	fmt.Printf("Starting User Service at port %d...\n", 8081)
	r.Run(":8081")
}
