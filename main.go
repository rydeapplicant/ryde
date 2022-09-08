package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	if err := Main(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Main() error {
	db, err := InitDb(os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("failed to initialize db: %w", err)
	}
	defer func() {
		if err := db.Disconnect(context.Background()); err != nil {
			panic(err)
		}
	}()

	router := InitControllers(db)

	return router.Run(":8080")
}

func InitDb(uri string) (*mongo.Client, error) {
	if uri == "" {
		return nil, fmt.Errorf("empty database URI")
	}

	ctx := context.Background()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return client, nil
}

func InitControllers(conn *mongo.Client) *gin.Engine {
	r := gin.Default()
	apis := r.Group("/apis")
	db := conn.Database("ryde")

	// User APIs
	userService := &userService{db.Collection("users")}
	userController := &UserController{userService}
	userRoutes(apis, userController)

	return r
}

func userRoutes(apis *gin.RouterGroup, userController *UserController) {
	apis.GET("/users/:id", userController.GetUser)
	apis.POST("/users", userController.CreateUser)
	apis.PUT("/users/:id", userController.UpdateUser)
	apis.DELETE("/users/:id", userController.DeleteUser)
}
