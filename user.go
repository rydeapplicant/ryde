package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	ID          *primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        *string             `json:"name,omitempty" bson:"name,omitempty"`
	Dob         *string             `json:"dob,omitempty" bson:"dob,omitempty"`
	Address     *string             `json:"address,omitempty" bson:"address,omitempty"`
	Description *string             `json:"description,omitempty" bson:"description,omitempty"`
	CreatedAt   *string             `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

type UserController struct {
	service UserService
}

func (u *UserController) GetUser(c *gin.Context) {
	user, err := u.service.get(c.Request.Context(), c.Param("id"))
	if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (u *UserController) CreateUser(c *gin.Context) {
	payload := struct {
		Name        *string `json:"name" binding:"required"`
		Dob         *string `json:"dob" binding:"required"`
		Address     *string `json:"address" binding:"required"`
		Description *string `json:"description"`
	}{}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid new user request"))
		return
	}

	user := &User{
		Name:        payload.Name,
		Dob:         payload.Dob,
		Address:     payload.Address,
		Description: payload.Description,
	}

	err := u.service.create(c.Request.Context(), user)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (u *UserController) UpdateUser(c *gin.Context) {
	payload := struct {
		Name        *string `json:"name" bson:"name,omitempty"`
		Dob         *string `json:"dob" bson:"dob,omitempty"`
		Address     *string `json:"address" bson:"address,omitempty"`
		Description *string `json:"description" bson:"description,omitempty"`
	}{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid update user request"))
		return
	}

	user := &User{
		Name:        payload.Name,
		Dob:         payload.Dob,
		Address:     payload.Address,
		Description: payload.Description,
	}
	updatedUser, err := u.service.update(c.Request.Context(), c.Param("id"), user)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if updatedUser == nil {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("user with id `%v` not found", c.Param("id")))
		return
	}

	c.JSON(http.StatusAccepted, updatedUser)
}

func (u *UserController) DeleteUser(c *gin.Context) {
	err := u.service.delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

type UserService interface {
	get(ctx context.Context, id string) (*User, error)
	create(ctx context.Context, user *User) error
	update(ctx context.Context, id string, user *User) (*User, error)
	delete(ctx context.Context, id string) error
}

type userService struct {
	coll *mongo.Collection
}

func (s *userService) get(ctx context.Context, id string) (*User, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var user User
	err = s.coll.FindOne(ctx, bson.D{{"_id", objectId}}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to find User: %w", err)
	}

	return &user, nil
}

func (s *userService) create(ctx context.Context, user *User) error {
	now := time.Now().UTC().Format(time.RFC3339)
	user.CreatedAt = &now

	res, err := s.coll.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	if insertedId, ok := res.InsertedID.(primitive.ObjectID); ok {
		user.ID = &insertedId
	}

	return nil
}

func (s *userService) update(ctx context.Context, id string, user *User) (*User, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user ID: %w", err)
	}

	res, err := s.coll.UpdateByID(ctx, objectId, bson.D{{Key: "$set", Value: user}})
	if err != nil {
		return nil, fmt.Errorf("failed to modify user with id %s: %w", id, err)
	}
	if res.MatchedCount < 1 {
		return nil, nil
	}

	return user, nil
}

func (s *userService) delete(ctx context.Context, id string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}

	if _, err = s.coll.DeleteOne(ctx, bson.D{{"_id", objectId}}); err != nil {
		return fmt.Errorf("failed to delete user with id %s: %w", id, err)
	}

	return nil
}
