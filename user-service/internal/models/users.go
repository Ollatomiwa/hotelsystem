package models

import (
	
)

type UserRole string 

const (
	RoleCustomer UserRole = "customer"
	RoleAdmin UserRole = "admin"
)

type User struct {
	Id string `json:"id"`
	Email string `json:"email"`
	PasswordHash string `json:"-"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Phone string `json:"phone,omitempty"`
	Role UserRole `json:"role"`
}

type CreateUserRequest struct {
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FirstName string `json:"firstName" binding:"required"`  
	LastName string `json:"lastName" binding:"required"` 
	Phone string `json:"phone,omitempty"`
}

type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"` 
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	User User `json:"user"`
}

type UpdateUserRequest struct {
	FirstName string `json:"firstName,omitempty" `  
	LastName string `json:"lasyName,omitempty"` 
	Phone string `json:"phone,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}