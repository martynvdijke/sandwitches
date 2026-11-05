// Package account owns whole-account operations: exporting everything the
// instance stores for a user and closing the account. The HTML settings form
// and the JSON API both go through this package so their behaviour cannot
// drift.
package account

import (
	"errors"
	"time"

	"github.com/martynvdijke/sandwitches-go/internal/database"
	"gorm.io/gorm"
)

var (
	// ErrNotFound means there is no account with the requested ID.
	ErrNotFound = errors.New("account not found")
	// ErrLastSuperuser means deleting the account would leave the instance
	// without any superuser able to administer it.
	ErrLastSuperuser = errors.New("the last superuser cannot be deleted")
	// ErrDemoAccount means the account is the shared public demo login. It is
	// deliberately not deletable, or the app-store reviewer login would vanish.
	ErrDemoAccount = errors.New("the demo account cannot be deleted")
)

// Service performs whole-account operations against a database handle.
type Service struct {
	DB *gorm.DB
}

// New builds a service. Pass database.DB in production; tests may pass their
// own handle.
func New(db *gorm.DB) *Service { return &Service{DB: db} }

// Delete closes an account in one transaction.
//
// The user row is removed, not merely hidden, so no email, password hash or
// OIDC subject survives. Content that other people may still depend on is kept
// but detached: authored recipes and order history stay with a NULL author, and
// ratings/comments, favorites, cart rows, group memberships and reset/mobile
// tokens are deleted. Orders themselves are retained because they are the
// instance's transaction record; their tracking tokens remain valid so a
// tracking link already shared keeps working, but the order no longer points at
// a person.
func (s *Service) Delete(userID uint) error {
	if s == nil || s.DB == nil || userID == 0 {
		return ErrNotFound
	}

	var user database.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	// The shared demo login must survive so reviewers can always sign in.
	if user.IsDemo {
		return ErrDemoAccount
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() { _ = tx.Rollback() }()

	// The guard is evaluated inside the transaction so two concurrent
	// deletions cannot both see "another superuser exists" and leave none.
	if user.IsSuperuser {
		var others int64
		if err := tx.Model(&database.User{}).
			Where("is_superuser = ? AND id <> ?", true, userID).
			Count(&others).Error; err != nil {
			return err
		}
		if others == 0 {
			return ErrLastSuperuser
		}
	}

	// Bump first: even if a later statement fails the transaction rolls back,
	// and on success any cookie still holding the old version is dead.
	if err := tx.Model(&database.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"session_version": gorm.Expr("CASE WHEN session_version < 1 THEN 2 ELSE session_version + 1 END"),
			"is_active":       false,
		}).Error; err != nil {
		return err
	}

	cleanup := []string{
		// Keep authored recipes and orders, but detach them from the person.
		"UPDATE recipes SET uploaded_by_id = NULL WHERE uploaded_by_id = ?",
		"UPDATE orders SET user_id = NULL WHERE user_id = ?",
		"UPDATE recipe_histories SET changed_by_id = NULL WHERE changed_by_id = ?",
		// Delete references owned by the account.
		"DELETE FROM user_groups WHERE user_id = ?",
		"DELETE FROM user_favorites WHERE user_id = ?",
		"DELETE FROM ratings WHERE user_id = ?",
		"DELETE FROM cart_items WHERE user_id = ?",
		"DELETE FROM password_reset_tokens WHERE user_id = ?",
		"DELETE FROM mobile_auth_codes WHERE user_id = ?",
	}
	for _, query := range cleanup {
		if err := tx.Exec(query, userID).Error; err != nil {
			return err
		}
	}

	if err := tx.Delete(&database.User{}, userID).Error; err != nil {
		return err
	}
	return tx.Commit().Error
}

// Export is the JSON document returned by GET /api/v1/me/export. It contains
// the user's own data only: no password hash, no session version, no reset
// token hashes, and never the settings singleton or another user's data.
type Export struct {
	ExportedAt  time.Time         `json:"exported_at"`
	Profile     ExportProfile     `json:"profile"`
	Groups      []string          `json:"groups"`
	Recipes     []ExportRecipe    `json:"recipes"`
	Ratings     []ExportRating    `json:"ratings"`
	Favorites   []ExportFavorite  `json:"favorites"`
	Cart        []ExportCartItem  `json:"cart"`
	Orders      []ExportOrder     `json:"orders"`
	RecipeEdits []ExportRecipeLog `json:"recipe_edits"`
}

type ExportProfile struct {
	ID          uint       `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Bio         string     `json:"bio"`
	Avatar      string     `json:"avatar"`
	Language    string     `json:"language"`
	Theme       string     `json:"theme"`
	IsStaff     bool       `json:"is_staff"`
	IsSuperuser bool       `json:"is_superuser"`
	OIDCSubject string     `json:"oidc_subject,omitempty"`
	DateJoined  time.Time  `json:"date_joined"`
	LastLogin   *time.Time `json:"last_login"`
	// Consent and privacy choices, so an export is a complete record of the
	// account's GDPR-relevant state.
	AcceptedTermsAt *time.Time `json:"accepted_terms_at"`
	AnalyticsOptOut bool       `json:"analytics_opt_out"`
}

type ExportRecipe struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description"`
	Ingredients  string    `json:"ingredients"`
	Instructions string    `json:"instructions"`
	Servings     int       `json:"servings"`
	Price        *float64  `json:"price"`
	Image        string    `json:"image"`
	IsApproved   bool      `json:"is_approved"`
	Tags         []string  `json:"tags"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ExportRating struct {
	RecipeID    uint      `json:"recipe_id"`
	RecipeTitle string    `json:"recipe_title"`
	Score       float64   `json:"score"`
	Comment     string    `json:"comment"`
	Image       string    `json:"image"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ExportFavorite struct {
	RecipeID uint   `json:"recipe_id"`
	Title    string `json:"title"`
	Slug     string `json:"slug"`
}

type ExportCartItem struct {
	RecipeID uint    `json:"recipe_id"`
	Title    string  `json:"title"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type ExportOrder struct {
	ID            uint              `json:"id"`
	Status        string            `json:"status"`
	Completed     bool              `json:"completed"`
	TotalPrice    float64           `json:"total_price"`
	TrackingToken string            `json:"tracking_token"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Items         []ExportOrderItem `json:"items"`
}

type ExportOrderItem struct {
	RecipeID uint    `json:"recipe_id"`
	Title    string  `json:"title"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type ExportRecipeLog struct {
	RecipeID  uint      `json:"recipe_id"`
	FieldName string    `json:"field_name"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	ChangedAt time.Time `json:"changed_at"`
}

// Export collects the user's own data. Tracking tokens are included because
// they belong to the user and are needed to keep order links working; they are
// not credentials for any other account.
func (s *Service) Export(userID uint) (*Export, error) {
	if s == nil || s.DB == nil || userID == 0 {
		return nil, ErrNotFound
	}

	var user database.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	out := &Export{
		ExportedAt:  time.Now().UTC(),
		Profile:     profileFromUser(&user),
		Groups:      []string{},
		Recipes:     []ExportRecipe{},
		Ratings:     []ExportRating{},
		Favorites:   []ExportFavorite{},
		Cart:        []ExportCartItem{},
		Orders:      []ExportOrder{},
		RecipeEdits: []ExportRecipeLog{},
	}

	// User does not declare a Groups association; read the join table directly.
	var groupNames []string
	if err := s.DB.Table("groups").
		Joins("JOIN user_groups ON user_groups.group_id = groups.id").
		Where("user_groups.user_id = ?", userID).
		Pluck("groups.name", &groupNames).Error; err != nil {
		return nil, err
	}
	if groupNames != nil {
		out.Groups = groupNames
	}

	var recipes []database.Recipe
	if err := s.DB.Preload("Tags").Where("uploaded_by_id = ?", userID).
		Order("created_at ASC").Find(&recipes).Error; err != nil {
		return nil, err
	}
	for _, r := range recipes {
		item := ExportRecipe{
			ID:           r.ID,
			Title:        r.Title,
			Slug:         r.Slug,
			Description:  r.Description,
			Ingredients:  r.Ingredients,
			Instructions: r.Instructions,
			Servings:     r.Servings,
			Price:        r.Price,
			Image:        r.Image,
			IsApproved:   r.IsApproved,
			Tags:         []string{},
			CreatedAt:    r.CreatedAt,
			UpdatedAt:    r.UpdatedAt,
		}
		for _, t := range r.Tags {
			item.Tags = append(item.Tags, t.Name)
		}
		out.Recipes = append(out.Recipes, item)
	}

	var ratings []database.Rating
	if err := s.DB.Preload("Recipe").Where("user_id = ?", userID).
		Order("created_at ASC").Find(&ratings).Error; err != nil {
		return nil, err
	}
	for _, r := range ratings {
		out.Ratings = append(out.Ratings, ExportRating{
			RecipeID:    r.RecipeID,
			RecipeTitle: r.Recipe.Title,
			Score:       r.Score,
			Comment:     r.Comment,
			Image:       r.Image,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		})
	}

	var favorites []database.Recipe
	if err := s.DB.Joins("JOIN user_favorites ON user_favorites.recipe_id = recipes.id").
		Where("user_favorites.user_id = ?", userID).
		Order("recipes.title ASC").Find(&favorites).Error; err != nil {
		return nil, err
	}
	for _, r := range favorites {
		out.Favorites = append(out.Favorites, ExportFavorite{RecipeID: r.ID, Title: r.Title, Slug: r.Slug})
	}

	var cart []database.CartItem
	if err := s.DB.Preload("Recipe").Where("user_id = ?", userID).
		Order("created_at ASC").Find(&cart).Error; err != nil {
		return nil, err
	}
	for _, ci := range cart {
		out.Cart = append(out.Cart, ExportCartItem{
			RecipeID: ci.RecipeID,
			Title:    ci.Recipe.Title,
			Quantity: ci.Quantity,
			Price:    ci.TotalPrice(),
		})
	}

	var orders []database.Order
	if err := s.DB.Preload("Items.Recipe").Where("user_id = ?", userID).
		Order("created_at ASC").Find(&orders).Error; err != nil {
		return nil, err
	}
	for _, o := range orders {
		item := ExportOrder{
			ID:            o.ID,
			Status:        o.Status,
			Completed:     o.Completed,
			TotalPrice:    o.TotalPrice,
			TrackingToken: o.TrackingToken,
			CreatedAt:     o.CreatedAt,
			UpdatedAt:     o.UpdatedAt,
			Items:         []ExportOrderItem{},
		}
		for _, oi := range o.Items {
			item.Items = append(item.Items, ExportOrderItem{
				RecipeID: oi.RecipeID,
				Title:    oi.Recipe.Title,
				Quantity: oi.Quantity,
				Price:    oi.Price,
			})
		}
		out.Orders = append(out.Orders, item)
	}

	var history []database.RecipeHistory
	if err := s.DB.Where("changed_by_id = ?", userID).
		Order("changed_at ASC").Find(&history).Error; err != nil {
		return nil, err
	}
	for _, h := range history {
		out.RecipeEdits = append(out.RecipeEdits, ExportRecipeLog{
			RecipeID:  h.RecipeID,
			FieldName: h.FieldName,
			OldValue:  h.OldValue,
			NewValue:  h.NewValue,
			ChangedAt: h.ChangedAt,
		})
	}

	return out, nil
}

func profileFromUser(u *database.User) ExportProfile {
	return ExportProfile{
		ID:              u.ID,
		Username:        u.Username,
		Email:           u.Email,
		FirstName:       u.FirstName,
		LastName:        u.LastName,
		Bio:             u.Bio,
		Avatar:          u.Avatar,
		Language:        u.Language,
		Theme:           u.Theme,
		IsStaff:         u.IsStaff,
		IsSuperuser:     u.IsSuperuser,
		OIDCSubject:     u.OIDCSub,
		DateJoined:      u.DateJoined,
		LastLogin:       u.LastLogin,
		AcceptedTermsAt: u.AcceptedTermsAt,
		AnalyticsOptOut: u.AnalyticsOptOut,
	}
}
