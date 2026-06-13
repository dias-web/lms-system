package dto

// LoginRequest is the body for POST /auth/login.
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"admin123"`
} // @name LoginRequest

// RefreshRequest is the body for POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOi..."`
} // @name RefreshRequest

// RegisterRequest is the body for POST /auth/register (admin only). Role must
// be one of the three realm roles.
type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=2,max=255" example:"jdoe"`
	Email     string `json:"email" binding:"required,email" example:"jdoe@lms.local"`
	Password  string `json:"password" binding:"required,min=6" example:"Str0ngPass"`
	FirstName string `json:"first_name" binding:"max=255" example:"John"`
	LastName  string `json:"last_name" binding:"max=255" example:"Doe"`
	Role      string `json:"role" binding:"required,oneof=ROLE_ADMIN ROLE_TEACHER ROLE_USER" example:"ROLE_USER"`
} // @name RegisterRequest

// UpdateProfileRequest is the body for PUT /auth/profile. All fields are
// optional; only provided ones are changed. There is deliberately no role
// field — users cannot change their own roles.
type UpdateProfileRequest struct {
	Email     string `json:"email" binding:"omitempty,email" example:"new@lms.local"`
	FirstName string `json:"first_name" binding:"omitempty,max=255" example:"John"`
	LastName  string `json:"last_name" binding:"omitempty,max=255" example:"Doe"`
} // @name UpdateProfileRequest

// ChangePasswordRequest is the body for PUT /auth/password. The current
// password is required and verified before the new one is applied.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"OldPass123"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"NewPass456"`
} // @name ChangePasswordRequest

// UserResponse describes a user managed via the admin endpoints.
type UserResponse struct {
	ID        string `json:"id" example:"b1c2d3e4-...."`
	Username  string `json:"username" example:"jdoe"`
	Email     string `json:"email" example:"jdoe@lms.local"`
	FirstName string `json:"first_name" example:"John"`
	LastName  string `json:"last_name" example:"Doe"`
	Role      string `json:"role" example:"ROLE_USER"`
} // @name UserResponse

// TokenResponse is the token pair returned by login and refresh endpoints.
// It mirrors the relevant fields of a Keycloak token response.
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in" example:"300"`
	RefreshExpiresIn int    `json:"refresh_expires_in" example:"604800"`
	TokenType        string `json:"token_type" example:"Bearer"`
} // @name TokenResponse