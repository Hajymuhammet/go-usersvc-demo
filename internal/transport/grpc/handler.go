package grpc

import (
	"context"
	"fmt"

	"go-usersvc-demo/internal/domain"
	"go-usersvc-demo/internal/service"
	"go-usersvc-demo/pkg/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler implements pb.UserServiceServer.
type Handler struct {
	pb.UnimplementedUserServiceServer
	svc *service.UserService
}

// NewHandler creates a new gRPC Handler.
func NewHandler(svc *service.UserService) *Handler {
	return &Handler{svc: svc}
}

func toProtoUser(u *domain.User) *pb.UserResponse {
	return &pb.UserResponse{
		Id:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// CreateUser creates a new user.
func (h *Handler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "name, email, and password are required")
	}

	user, err := h.svc.CreateUser(ctx, domain.CreateUserInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if err.Error() == "email already registered" {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("could not create user: %v", err))
	}

	return toProtoUser(user), nil
}

// GetUserByID fetches a user by ID.
func (h *Handler) GetUserByID(ctx context.Context, req *pb.GetUserByIDRequest) (*pb.UserResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	user, err := h.svc.GetUserByID(ctx, req.Id)
	if err != nil {
		if err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "could not fetch user")
	}

	return toProtoUser(user), nil
}

// ListUsers returns a paginated list of users.
func (h *Handler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	page := int(req.Page)
	limit := int(req.Limit)
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	result, err := h.svc.ListUsers(ctx, domain.ListFilter{Page: page, Limit: limit})
	if err != nil {
		return nil, status.Error(codes.Internal, "could not list users")
	}

	resp := &pb.ListUsersResponse{
		Total: result.Total,
		Page:  int32(result.Page),
		Limit: int32(result.Limit),
	}
	for _, u := range result.Data {
		resp.Data = append(resp.Data, &pb.User{
			Id:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return resp, nil
}

// UpdateUser updates an existing user.
func (h *Handler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	input := domain.UpdateUserInput{}
	if req.Name != "" {
		input.Name = &req.Name
	}
	if req.Email != "" {
		input.Email = &req.Email
	}
	if req.Password != "" {
		input.Password = &req.Password
	}

	user, err := h.svc.UpdateUser(ctx, req.Id, input)
	if err != nil {
		if err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "could not update user")
	}

	return toProtoUser(user), nil
}

// DeleteUser removes a user by ID.
func (h *Handler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := h.svc.DeleteUser(ctx, req.Id); err != nil {
		if err.Error() == "user not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "could not delete user")
	}

	return &pb.DeleteUserResponse{Success: true, Message: "user deleted successfully"}, nil
}
