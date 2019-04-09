package model

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// NotFound indicates that given resource was not found in database
	NotFound = status.New(codes.NotFound, "Account not found")
	// AlreadyExists indicates that resource with same credentials already exists
	AlreadyExists = status.New(codes.AlreadyExists, "Account already exists")
	// InvalidCode indicates that provided activation code is invalid
	InvalidCode = status.New(codes.InvalidArgument, "Invalid activation code")
	// Unauthenticated indicates that provided credentials are invalid
	Unauthenticated = status.New(codes.Unauthenticated, "Invalid credentials")
	// Unavailable indicates that downstream operation failed
	Unavailable = status.New(codes.Unavailable, "Temporarily unavailable")
	// Internal indicates that internal error occured
	Internal = status.New(codes.Internal, "Internal error")
)
