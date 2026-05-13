package http

import (
	"net/http"
	"strconv"

	"go-usersvc-demo/internal/domain"
	"go-usersvc-demo/internal/pkg/httputility"
	"go-usersvc-demo/internal/pkg/logger"
	"go-usersvc-demo/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type Handler struct {
	svc          *service.UserService
	authSvc      *service.AuthService
	emailSvc     *service.EmailService
}

func NewHandler(svc *service.UserService, authSvc *service.AuthService, emailSvc *service.EmailService) *Handler {
	return &Handler{svc: svc, authSvc: authSvc, emailSvc: emailSvc}
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

type sendEmailRequest struct {
	To       string `json:"to" validate:"required,email"`
	Template string `json:"template" validate:"required,oneof=welcome reset verification"`
	Data     map[string]string `json:"data" validate:"required"`
}

type sendEmailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
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
		httputility.RespondError(c, domain.NewValidationError("invalid request body", err.Error()))
		return
	}
	if err := validate.Struct(input); err != nil {
		httputility.RespondError(c, domain.NewValidationError("validation failed", err.Error()))
		return
	}

	user, err := h.svc.CreateUser(c.Request.Context(), input)
	if err != nil {
		httputility.RespondError(c, err)
		return
	}

	httputility.RespondSuccess(c, http.StatusCreated, user)
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
		httputility.RespondError(c, domain.NewValidationError("invalid request body", err.Error()))
		return
	}
	if err := validate.Struct(input); err != nil {
		httputility.RespondError(c, domain.NewValidationError("validation failed", err.Error()))
		return
	}

	accessToken, refreshToken, _, err := h.authSvc.Login(c.Request.Context(), input.Email, input.Password)
	if err != nil {
		httputility.RespondError(c, err)
		return
	}

	httputility.RespondSuccess(c, http.StatusOK, tokenResponse{AccessToken: accessToken, RefreshToken: refreshToken})
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
		httputility.RespondError(c, domain.NewValidationError("invalid request body", err.Error()))
		return
	}
	if err := validate.Struct(input); err != nil {
		httputility.RespondError(c, domain.NewValidationError("validation failed", err.Error()))
		return
	}

	accessToken, refreshToken, err := h.authSvc.Refresh(c.Request.Context(), input.RefreshToken)
	if err != nil {
		httputility.RespondError(c, err)
		return
	}

	httputility.RespondSuccess(c, http.StatusOK, tokenResponse{AccessToken: accessToken, RefreshToken: refreshToken})
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
// @Security     BearerAuth
// @Router       /users/{id} [get]
func (h *Handler) GetUserByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		httputility.RespondError(c, domain.NewValidationError("invalid user id", ""))
		return
	}

	user, err := h.svc.GetUserByID(c.Request.Context(), id)
	if err != nil {
		httputility.RespondError(c, err)
		return
	}
	user.Password = "" // never expose password hash

	httputility.RespondSuccess(c, http.StatusOK, user)
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
// @Security     BearerAuth
// @Router       /users/me [get]
func (h *Handler) GetProfile(c *gin.Context) {
	userID, ok := UserIDFromContext(c.Request.Context())
	if !ok {
		httputility.RespondError(c, domain.NewUnauthorizedError("unauthorized"))
		return
	}

	user, err := h.svc.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		httputility.RespondError(c, err)
		return
	}
	user.Password = ""

	httputility.RespondSuccess(c, http.StatusOK, user)
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
// @Security     BearerAuth
// @Router       /users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	result, err := h.svc.ListUsers(c.Request.Context(), domain.ListFilter{Page: page, Limit: limit})
	if err != nil {
		httputility.RespondError(c, domain.NewInternalError("failed to list users", err))
		return
	}

	httputility.RespondSuccess(c, http.StatusOK, result)
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
// @Security     BearerAuth
// @Router       /users/{id} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		httputility.RespondError(c, domain.NewValidationError("invalid user id", ""))
		return
	}

	var input domain.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		httputility.RespondError(c, domain.NewValidationError("invalid request body", err.Error()))
		return
	}
	if err := validate.Struct(input); err != nil {
		httputility.RespondError(c, domain.NewValidationError("validation failed", err.Error()))
		return
	}

	user, err := h.svc.UpdateUser(c.Request.Context(), id, input)
	if err != nil {
		httputility.RespondError(c, err)
		return
	}
	user.Password = ""

	httputility.RespondSuccess(c, http.StatusOK, user)
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
// @Security     BearerAuth
// @Router       /users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		httputility.RespondError(c, domain.NewValidationError("invalid user id", ""))
		return
	}

	if err := h.svc.DeleteUser(c.Request.Context(), id); err != nil {
		httputility.RespondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// SendEmail godoc
// @Summary      Send email
// @Description  Send an email to a recipient with specified template and data
// @Tags         email
// @Accept       json
// @Produce      json
// @Param        email  body      sendEmailRequest   true  "Email request"
// @Success      200    {object}  sendEmailResponse
// @Failure      400    {object}  map[string]string
// @Failure      401    {object}  map[string]string
// @Failure      422    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Security     BearerAuth
// @Router       /email/send [post]
func (h *Handler) SendEmail(c *gin.Context) {
	userID, ok := UserIDFromContext(c.Request.Context())
	if !ok {
		httputility.RespondError(c, domain.NewUnauthorizedError("unauthorized"))
		return
	}

	var input sendEmailRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		httputility.RespondError(c, domain.NewValidationError("invalid request body", err.Error()))
		return
	}

	if err := validate.Struct(input); err != nil {
		httputility.RespondError(c, domain.NewValidationError("validation failed", err.Error()))
		return
	}

	logger.Get().Info("email request",
		"user_id", userID,
		"recipient", input.To,
		"template", input.Template,
	)

	switch input.Template {
	case "welcome":
		name, ok := input.Data["name"]
		if !ok {
			httputility.RespondError(c, domain.NewValidationError("missing required data", "name"))
			return
		}
		if err := h.emailSvc.SendWelcomeEmail(c.Request.Context(), input.To, name); err != nil {
			logger.Get().Error("failed to send welcome email", "error", err)
			httputility.RespondError(c, domain.NewInternalError("failed to send email", err))
			return
		}

	case "reset":
		resetLink, ok := input.Data["reset_link"]
		if !ok {
			httputility.RespondError(c, domain.NewValidationError("missing required data", "reset_link"))
			return
		}
		if err := h.emailSvc.SendPasswordResetEmail(c.Request.Context(), input.To, resetLink); err != nil {
			logger.Get().Error("failed to send reset email", "error", err)
			httputility.RespondError(c, domain.NewInternalError("failed to send email", err))
			return
		}

	case "verification":
		verificationLink, ok := input.Data["verification_link"]
		if !ok {
			httputility.RespondError(c, domain.NewValidationError("missing required data", "verification_link"))
			return
		}
		if err := h.emailSvc.SendVerificationEmail(c.Request.Context(), input.To, verificationLink); err != nil {
			logger.Get().Error("failed to send verification email", "error", err)
			httputility.RespondError(c, domain.NewInternalError("failed to send email", err))
			return
		}

	default:
		httputility.RespondError(c, domain.NewValidationError("unknown template type", input.Template))
		return
	}

	httputility.RespondSuccess(c, http.StatusOK, sendEmailResponse{
		Success: true,
		Message: "Email sent successfully",
	})
}
