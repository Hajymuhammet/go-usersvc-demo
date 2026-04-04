package http

import (
	"net/http"
	"strconv"

	"go-usersvc-demo/internal/domain"
	"go-usersvc-demo/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Handler holds the HTTP handlers for user operations.
type Handler struct {
	svc     *service.UserService
	authSvc *service.AuthService
}

// NewHandler creates a new HTTP Handler.
func NewHandler(svc *service.UserService, authSvc *service.AuthService) *Handler {
	return &Handler{svc: svc, authSvc: authSvc}
}

func respondError(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"error": msg})
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// CreateUser godoc
// @Summary      Create a new user
// @Description  Create a new user with name, email and password
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      domain.CreateUserInput  true  "User Create Input"
// @Success      201   {object}  domain.User
// @Failure      400   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      422   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var input domain.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validate.Struct(input); err != nil {
		respondError(c, http.StatusUnprocessableEntity, err.Error())
		return
	}

	user, err := h.svc.CreateUser(c.Request.Context(), input)
	if err != nil {
		if err.Error() == "email already registered" {
			respondError(c, http.StatusConflict, err.Error())
			return
		}
		respondError(c, http.StatusInternalServerError, "could not create user")
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login godoc
// @Summary      Login and receive access/refresh tokens
// @Description  Authenticate user credentials and return JWT access and refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      loginRequest  true  "Login credentials"
// @Success      200          {object}  tokenResponse
// @Failure      400          {object}  map[string]string
// @Failure      401          {object}  map[string]string
// @Failure      500          {object}  map[string]string
// @Router       /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var input loginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validate.Struct(input); err != nil {
		respondError(c, http.StatusUnprocessableEntity, err.Error())
		return
	}

	accessToken, refreshToken, _, err := h.authSvc.Login(c.Request.Context(), input.Email, input.Password)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	c.JSON(http.StatusOK, tokenResponse{AccessToken: accessToken, RefreshToken: refreshToken})
}

// Refresh godoc
// @Summary      Refresh access token
// @Description  Use a refresh token to issue a new access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        refresh  body      refreshRequest  true  "Refresh token"
// @Success      200      {object}  tokenResponse
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var input refreshRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validate.Struct(input); err != nil {
		respondError(c, http.StatusUnprocessableEntity, err.Error())
		return
	}

	accessToken, refreshToken, err := h.authSvc.Refresh(c.Request.Context(), input.RefreshToken)
	if err != nil {
		respondError(c, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	c.JSON(http.StatusOK, tokenResponse{AccessToken: accessToken, RefreshToken: refreshToken})
}

// GetUserByID godoc
// @Summary      Get user by ID
// @Description  Retrieve a user's details by their database ID
// @Tags         users
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  domain.User
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/{id} [get]
func (h *Handler) GetUserByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.svc.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			respondError(c, http.StatusNotFound, err.Error())
			return
		}
		respondError(c, http.StatusInternalServerError, "could not fetch user")
		return
	}
	user.Password = "" // never expose password hash

	c.JSON(http.StatusOK, user)
}

// GetProfile godoc
// @Summary      Get current user profile
// @Description  Get profile data for the authenticated user
// @Tags         users
// @Produce      json
// @Success      200  {object}  domain.User
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/me [get]
func (h *Handler) GetProfile(c *gin.Context) {
	userID, ok := UserIDFromContext(c.Request.Context())
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.svc.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			respondError(c, http.StatusNotFound, err.Error())
			return
		}
		respondError(c, http.StatusInternalServerError, "could not fetch profile")
		return
	}
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// ListUsers godoc
// @Summary      List users
// @Description  Get a paginated list of users
// @Tags         users
// @Produce      json
// @Param        page   query     int  false  "Page number"
// @Param        limit  query     int  false  "Items per page"
// @Success      200    {object}  domain.UserList
// @Failure      500    {object}  map[string]string
// @Router       /users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	result, err := h.svc.ListUsers(c.Request.Context(), domain.ListFilter{Page: page, Limit: limit})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "could not list users")
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateUser godoc
// @Summary      Update user
// @Description  Update allowed fields of an existing user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      int                     true  "User ID"
// @Param        user  body      domain.UpdateUserInput  true  "User Update Input"
// @Success      200   {object}  domain.User
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      422   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users/{id} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var input domain.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validate.Struct(input); err != nil {
		respondError(c, http.StatusUnprocessableEntity, err.Error())
		return
	}

	user, err := h.svc.UpdateUser(c.Request.Context(), id, input)
	if err != nil {
		if err.Error() == "user not found" {
			respondError(c, http.StatusNotFound, err.Error())
			return
		}
		respondError(c, http.StatusInternalServerError, "could not update user")
		return
	}
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// DeleteUser godoc
// @Summary      Delete user
// @Description  Remove a user from the system by ID
// @Tags         users
// @Param        id   path      int  true  "User ID"
// @Success      204  {string}  string "No Content"
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.svc.DeleteUser(c.Request.Context(), id); err != nil {
		if err.Error() == "user not found" {
			respondError(c, http.StatusNotFound, err.Error())
			return
		}
		respondError(c, http.StatusInternalServerError, "could not delete user")
		return
	}

	c.Status(http.StatusNoContent)
}
