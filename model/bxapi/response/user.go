package response

import "github.com/mo7amed4522/tayara_api/model/bxapi"

// UserLoginResponse returned on successful user login (token + profile).
type UserLoginResponse struct {
	User      bxapi.BxsUserProfile `json:"user"`
	Token     string               `json:"token"`
	ExpiresAt int64                `json:"expiresAt"`
}

// GetProfileData return on success user profile with favoiute , past and upcoming trip
type BxGetUserProfileData struct {
	User      bxapi.BxsUserProfile `json:"user"`
	PastTrips int32                `json:"pastTrips"`
	Upcoming  int32                `json:"upcoming"`
	Favorites int32                `json:"favorites"`
}

// UpdateProfileData return user data
type BxUpdateUserProfileData struct {
	User bxapi.BxsUserProfile `json:"user"`
}
