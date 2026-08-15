package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/martynvdijke/sandwitches-go/internal/database"
)

// SettingDTO mirrors the v2.x SettingSchema. All fields snake_case.
type SettingDTO struct {
	SiteName          string `json:"site_name"`
	SiteDescription   string `json:"site_description"`
	Email             string `json:"email"`
	AIConnectionPoint string `json:"ai_connection_point"`
	AIModel           string `json:"ai_model"`
	AIAPIKey          string `json:"ai_api_key"`
	LogLevel          string `json:"log_level"`
	GotifyURL         string `json:"gotify_url"`
	GotifyToken       string `json:"gotify_token"`
	OTelEndpoint      string `json:"otel_endpoint"`
	OTelEnabled       bool   `json:"otel_enabled"`
}

func settingToDTO(s *database.Setting) SettingDTO {
	return SettingDTO{
		SiteName:          s.SiteName,
		SiteDescription:   s.SiteDescription,
		Email:             s.Email,
		AIConnectionPoint: s.AIConnectionPoint,
		AIModel:           s.AIModel,
		AIAPIKey:          s.AIAPIKey,
		LogLevel:          s.LogLevel,
		GotifyURL:         s.GotifyURL,
		GotifyToken:       s.GotifyToken,
		OTelEndpoint:      s.OTelEndpoint,
		OTelEnabled:       s.OTelEnabled,
	}
}

// SettingUpdateInput accepts partial snake_case updates (pointer fields).
type SettingUpdateInput struct {
	SiteName          *string `json:"site_name"`
	SiteDescription   *string `json:"site_description"`
	Email             *string `json:"email"`
	AIConnectionPoint *string `json:"ai_connection_point"`
	AIModel           *string `json:"ai_model"`
	AIAPIKey          *string `json:"ai_api_key"`
	LogLevel          *string `json:"log_level"`
	GotifyURL         *string `json:"gotify_url"`
	GotifyToken       *string `json:"gotify_token"`
	OTelEndpoint      *string `json:"otel_endpoint"`
	OTelEnabled       *bool   `json:"otel_enabled"`
}

func applySettingUpdates(s *database.Setting, in SettingUpdateInput) {
	if in.SiteName != nil {
		s.SiteName = *in.SiteName
	}
	if in.SiteDescription != nil {
		s.SiteDescription = *in.SiteDescription
	}
	if in.Email != nil {
		s.Email = *in.Email
	}
	if in.AIConnectionPoint != nil {
		s.AIConnectionPoint = *in.AIConnectionPoint
	}
	if in.AIModel != nil {
		s.AIModel = *in.AIModel
	}
	if in.AIAPIKey != nil {
		s.AIAPIKey = *in.AIAPIKey
	}
	if in.LogLevel != nil {
		s.LogLevel = *in.LogLevel
	}
	if in.GotifyURL != nil {
		s.GotifyURL = *in.GotifyURL
	}
	if in.GotifyToken != nil {
		s.GotifyToken = *in.GotifyToken
	}
	if in.OTelEndpoint != nil {
		s.OTelEndpoint = *in.OTelEndpoint
	}
	if in.OTelEnabled != nil {
		s.OTelEnabled = *in.OTelEnabled
	}
}

// UserDTO mirrors the v2.x UserSchema: excludes email, password and the
// staff/superuser/active/joined bookkeeping fields.
type UserDTO struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Avatar    string `json:"avatar"`
	Bio       string `json:"bio"`
	Language  string `json:"language"`
	Theme     string `json:"theme"`
}

func userToDTO(u *database.User) UserDTO {
	return UserDTO{
		ID:        u.ID,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Avatar:    u.Avatar,
		Bio:       u.Bio,
		Language:  u.Language,
		Theme:     u.Theme,
	}
}

func usersToDTOs(users []database.User) []UserDTO {
	out := make([]UserDTO, len(users))
	for i, u := range users {
		out[i] = userToDTO(&u)
	}
	return out
}

// UserPublicDTO mirrors the v2.x UserPublicSchema, used inside recipes
// (favorited_by) and ratings (user).
type UserPublicDTO struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Avatar    string `json:"avatar"`
}

func userPublicToDTO(u *database.User) UserPublicDTO {
	return UserPublicDTO{
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Avatar:    u.Avatar,
	}
}

// TagDTO mirrors the v2.x TagSchema.
type TagDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func tagToDTO(t *database.Tag) TagDTO {
	return TagDTO{ID: t.ID, Name: t.Name, Slug: t.Slug}
}

func tagsToDTOs(tags []database.Tag) []TagDTO {
	out := make([]TagDTO, len(tags))
	for i, t := range tags {
		out[i] = tagToDTO(&t)
	}
	return out
}

// OrderDTO mirrors the v2.x OrderSchema (list/create responses). Internal
// fields such as tracking_token, user_id and completed are never exposed.
type OrderDTO struct {
	ID         uint      `json:"id"`
	Status     string    `json:"status"`
	TotalPrice float64   `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// OrderDetailDTO mirrors the v2.x OrderDetailSchema (single-order responses).
type OrderDetailDTO struct {
	OrderDTO
	Items []OrderItemDTO `json:"items"`
}

// OrderItemDTO mirrors the v2.x OrderItemSchema.
type OrderItemDTO struct {
	ID       uint    `json:"id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Recipe   gin.H   `json:"recipe"`
}

func orderToDTO(o *database.Order) OrderDTO {
	return OrderDTO{
		ID:         o.ID,
		Status:     o.Status,
		TotalPrice: o.TotalPrice,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,
	}
}

func orderToDetailDTO(o *database.Order) OrderDetailDTO {
	items := make([]OrderItemDTO, len(o.Items))
	for i, it := range o.Items {
		items[i] = OrderItemDTO{
			ID:       it.ID,
			Quantity: it.Quantity,
			Price:    it.Price,
			Recipe:   recipeToJSON(&it.Recipe),
		}
	}
	return OrderDetailDTO{
		OrderDTO: orderToDTO(o),
		Items:    items,
	}
}

func ordersToDTOs(orders []database.Order) []OrderDTO {
	out := make([]OrderDTO, len(orders))
	for i, o := range orders {
		out[i] = orderToDTO(&o)
	}
	return out
}

// CartItemDTO mirrors the v2.x CartItemSchema.
type CartItemDTO struct {
	ID        uint      `json:"id"`
	Recipe    gin.H     `json:"recipe"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func cartItemToDTO(ci *database.CartItem) CartItemDTO {
	return CartItemDTO{
		ID:        ci.ID,
		Recipe:    recipeToJSON(&ci.Recipe),
		Quantity:  ci.Quantity,
		CreatedAt: ci.CreatedAt,
		UpdatedAt: ci.UpdatedAt,
	}
}

func cartItemsToDTOs(items []database.CartItem) []CartItemDTO {
	out := make([]CartItemDTO, len(items))
	for i, it := range items {
		out[i] = cartItemToDTO(&it)
	}
	return out
}

// RatingDTO mirrors the v2.x RatingSchema (fields __all__ + user).
type RatingDTO struct {
	ID        uint          `json:"id"`
	Recipe    uint          `json:"recipe"`
	Score     float64       `json:"score"`
	Comment   string        `json:"comment"`
	User      UserPublicDTO `json:"user"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

func ratingToDTO(r *database.Rating) RatingDTO {
	user := UserPublicDTO{}
	if r.User.ID != 0 || r.User.Username != "" {
		user = userPublicToDTO(&r.User)
	}
	return RatingDTO{
		ID:        r.ID,
		Recipe:    r.RecipeID,
		Score:     r.Score,
		Comment:   r.Comment,
		User:      user,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
