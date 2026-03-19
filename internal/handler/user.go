// Package handler implements the HTTP handlers for user-svc.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luciocarvalhojr/observatory-user-svc/internal/domain"
	"github.com/luciocarvalhojr/observatory-user-svc/internal/publisher"
	"github.com/luciocarvalhojr/observatory-user-svc/internal/repository"
	"github.com/rs/zerolog/log"
)

// User handles all user CRUD endpoints.
type User struct {
	repo      repository.Repository
	publisher *publisher.Publisher
}

// NewUser creates a new User handler.
func NewUser(repo repository.Repository, pub *publisher.Publisher) *User {
	return &User{repo: repo, publisher: pub}
}

// Register wires all user routes onto the given router group.
func (u *User) Register(rg *gin.RouterGroup) {
	rg.POST("", u.Create)
	rg.GET("/:id", u.Get)
	rg.PUT("/:id", u.Update)
	rg.DELETE("/:id", u.Delete)
}

// Create godoc
// @Summary      Create a user
// @Description  Creates a new user and publishes user.created to NATS
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      domain.CreateUserRequest  true  "User payload"
// @Success      201   {object}  domain.User
// @Failure      400   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users [post]
func (u *User) Create(c *gin.Context) {
	var req domain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := u.repo.Create(c.Request.Context(), &req)
	if err != nil {
		log.Error().Err(err).Msg("create user failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if err := u.publisher.UserCreated(c.Request.Context(), user); err != nil {
		log.Error().Err(err).Str("user_id", user.ID).Msg("failed to publish user.created")
		// non-fatal: user was created, event publish failure is logged only
	}

	log.Info().Str("user_id", user.ID).Str("email", user.Email).Msg("user created")
	c.JSON(http.StatusCreated, user)
}

// Get godoc
// @Summary      Get a user
// @Description  Returns a user by ID
// @Tags         users
// @Produce      json
// @Param        id   path      string  true  "User ID"
// @Success      200  {object}  domain.User
// @Failure      404  {object}  map[string]string
// @Router       /users/{id} [get]
func (u *User) Get(c *gin.Context) {
	user, err := u.repo.GetByID(c.Request.Context(), c.Param("id"))
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("get user failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// Update godoc
// @Summary      Update a user
// @Description  Updates a user's name
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      string                    true  "User ID"
// @Param        body  body      domain.UpdateUserRequest  true  "Update payload"
// @Success      200   {object}  domain.User
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /users/{id} [put]
func (u *User) Update(c *gin.Context) {
	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := u.repo.Update(c.Request.Context(), c.Param("id"), &req)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("update user failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	log.Info().Str("user_id", user.ID).Msg("user updated")
	c.JSON(http.StatusOK, user)
}

// Delete godoc
// @Summary      Delete a user
// @Description  Deletes a user and publishes user.deleted to NATS
// @Tags         users
// @Param        id   path  string  true  "User ID"
// @Success      204
// @Failure      404  {object}  map[string]string
// @Router       /users/{id} [delete]
func (u *User) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := u.repo.Delete(c.Request.Context(), id); errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	} else if err != nil {
		log.Error().Err(err).Msg("delete user failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if err := u.publisher.UserDeleted(c.Request.Context(), id); err != nil {
		log.Error().Err(err).Str("user_id", id).Msg("failed to publish user.deleted")
	}

	log.Info().Str("user_id", id).Msg("user deleted")
	c.Status(http.StatusNoContent)
}
