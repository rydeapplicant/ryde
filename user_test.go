package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MockDb struct {
	mock.Mock
}

type MockCollection struct {
	mock.Mock
}

var db = new(MockDb)
var coll = new(MockCollection)

func setup() {
	db.On("Database", mock.Anything).Return(coll)
}

func stringPtr(s string) *string {
	return &s
}

type MockUserService struct {
	mockGet    func() (*User, error)
	mockCreate func() error
	mockUpdate func() (*User, error)
	mockDelete func() error
}

func (m MockUserService) get(context.Context, string) (*User, error) {
	return m.mockGet()
}

func (m MockUserService) create(context.Context, *User) error {
	return m.mockCreate()
}

func (m MockUserService) update(context.Context, string, *User) (*User, error) {
	return m.mockUpdate()
}

func (m MockUserService) delete(context.Context, string) error {
	return m.mockDelete()
}

func NewMockUserService() MockUserService {
	return MockUserService{
		mockGet: func() (*User, error) {
			return nil, nil
		},
		mockCreate: func() error {
			return nil
		},
		mockUpdate: func() (*User, error) {
			return nil, nil
		},
		mockDelete: func() error {
			return nil
		},
	}
}

func TestUserControllerGetUser(t *testing.T) {
	setup()

	objId := primitive.NewObjectID()
	id := objId.Hex()
	user := &User{
		ID:          &objId,
		Name:        stringPtr("John Doe"),
		Dob:         stringPtr("1/1/2022"),
		Address:     stringPtr("1 Singapore Road"),
		Description: stringPtr("test user"),
		CreatedAt:   stringPtr("now"),
	}

	r := gin.Default()
	svc := NewMockUserService()
	svc.mockGet = func() (*User, error) {
		return user, nil
	}
	userRoutes(r.Group("/apis"), &UserController{svc})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/apis/users/"+id, nil)
	r.ServeHTTP(w, req)

	var got User
	json.Unmarshal(w.Body.Bytes(), &got)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "John Doe", *got.Name)
	assert.Equal(t, "1/1/2022", *got.Dob)
	assert.Equal(t, "1 Singapore Road", *got.Address)
	assert.Equal(t, "test user", *got.Description)
	assert.Equal(t, "now", *got.CreatedAt)
}

func TestUserControllerGetUserNotFound(t *testing.T) {
	setup()
	r := gin.Default()
	svc := NewMockUserService()
	svc.mockGet = func() (*User, error) {
		return nil, mongo.ErrNoDocuments
	}
	userRoutes(r.Group("/apis"), &UserController{svc})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/apis/users/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestUserControllerCreateUser(t *testing.T) {
	setup()
	r := gin.Default()
	svc := NewMockUserService()
	userRoutes(r.Group("/apis"), &UserController{svc})

	payload := struct {
		Name        string `json:"name" binding:"required"`
		Dob         string `json:"dob" binding:"required"`
		Address     string `json:"address" binding:"required"`
		Description string `json:"description"`
	}{
		Name:        "John Doe",
		Dob:         "1/2/3",
		Address:     "1 Singapore Road",
		Description: "test create",
	}

	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/apis/users", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	var got User
	json.Unmarshal(w.Body.Bytes(), &got)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "John Doe", *got.Name)
	assert.Equal(t, "1/2/3", *got.Dob)
	assert.Equal(t, "1 Singapore Road", *got.Address)
	assert.Equal(t, "test create", *got.Description)
}

func TestUserControllerCreateUserErr(t *testing.T) {
	setup()
	r := gin.Default()
	svc := NewMockUserService()
	svc.mockCreate = func() error {
		return fmt.Errorf("oops")
	}
	userRoutes(r.Group("/apis"), &UserController{svc})

	payload := struct {
		Name        string `json:"name" binding:"required"`
		Dob         string `json:"dob" binding:"required"`
		Address     string `json:"address" binding:"required"`
		Description string `json:"description"`
	}{
		Name:        "John Doe",
		Dob:         "1/2/3",
		Address:     "1 Singapore Road",
		Description: "test create",
	}

	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/apis/users", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserControllerUpdateUser(t *testing.T) {
	setup()
	r := gin.Default()
	svc := NewMockUserService()

	objId := primitive.NewObjectID()
	id := objId.Hex()

	updatedUser := &User{
		ID:          &objId,
		Name:        stringPtr("John Doe"),
		Dob:         stringPtr("1/1/2022"),
		Address:     stringPtr("1 Singapore Road"),
		Description: stringPtr("test update user"),
	}
	svc.mockUpdate = func() (*User, error) {
		return updatedUser, nil
	}
	userRoutes(r.Group("/apis"), &UserController{svc})

	payload := struct {
		Name        string `json:"name" bson:"name,omitempty"`
		Dob         string `json:"dob" bson:"dob,omitempty"`
		Address     string `json:"address" bson:"address,omitempty"`
		Description string `json:"description" bson:"description,omitempty"`
	}{
		Name:        "John Doe",
		Dob:         "1/1/2022",
		Address:     "1 Singapore Road",
		Description: "test update user",
	}

	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/apis/users/"+id, bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	var got User
	json.Unmarshal(w.Body.Bytes(), &got)

	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Equal(t, "John Doe", *got.Name)
	assert.Equal(t, "1/1/2022", *got.Dob)
	assert.Equal(t, "1 Singapore Road", *got.Address)
	assert.Equal(t, "test update user", *got.Description)
}

func TestUserControllerUpdateUserNotFound(t *testing.T) {
	setup()
	r := gin.Default()
	svc := NewMockUserService()
	userRoutes(r.Group("/apis"), &UserController{svc})

	payload := struct {
		Name        string `json:"name" bson:"name,omitempty"`
		Dob         string `json:"dob" bson:"dob,omitempty"`
		Address     string `json:"address" bson:"address,omitempty"`
		Description string `json:"description" bson:"description,omitempty"`
	}{
		Name:        "John Doe",
		Dob:         "1/1/2022",
		Address:     "1 Singapore Road",
		Description: "test update user",
	}

	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/apis/users/1", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserControllerUpdateUserErr(t *testing.T) {
	setup()
	r := gin.Default()
	svc := NewMockUserService()
	svc.mockUpdate = func() (*User, error) {
		return nil, fmt.Errorf("oops")
	}
	userRoutes(r.Group("/apis"), &UserController{svc})

	payload := struct {
		Name        string `json:"name" bson:"name,omitempty"`
		Dob         string `json:"dob" bson:"dob,omitempty"`
		Address     string `json:"address" bson:"address,omitempty"`
		Description string `json:"description" bson:"description,omitempty"`
	}{
		Name:        "John Doe",
		Dob:         "1/1/2022",
		Address:     "1 Singapore Road",
		Description: "test update user",
	}

	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/apis/users/1", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserControllerDeleteUser(t *testing.T) {
	setup()
	r := gin.Default()
	svc := NewMockUserService()
	userRoutes(r.Group("/apis"), &UserController{svc})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/apis/users/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserControllerDeleteUserErr(t *testing.T) {
	setup()
	r := gin.Default()
	svc := NewMockUserService()
	svc.mockDelete = func() error {
		return fmt.Errorf("oops")
	}
	userRoutes(r.Group("/apis"), &UserController{svc})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/apis/users/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
