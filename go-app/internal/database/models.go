package database

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Setting is a singleton model (only one row).
type Setting struct {
	ID               uint   `gorm:"primaryKey"`
	SiteName         string `gorm:"size:255;default:Sandwitches"`
	SiteDescription  string
	Email            string
	AIConnectionPoint string
	AIModel          string
	AIAPIKey         string
	LogLevel         string `gorm:"size:10;default:INFO"`
	GotifyURL        string
	GotifyToken      string
}

func (Setting) TableName() string { return "settings" }

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Username     string    `gorm:"size:150;uniqueIndex;not null"`
	Email        string    `gorm:"size:254"`
	Password     string    `gorm:"size:128;not null"`
	FirstName    string    `gorm:"size:150"`
	LastName     string    `gorm:"size:150"`
	Avatar       string
	Bio          string
	Language     string `gorm:"size:10;default:en"`
	Theme        string `gorm:"size:10;default:light"`
	IsStaff      bool   `gorm:"default:false"`
	IsActive     bool   `gorm:"default:true"`
	IsSuperuser  bool   `gorm:"default:false"`
	DateJoined   time.Time
	LastLogin    *time.Time
	Favorites    []Recipe  `gorm:"many2many:user_favorites;"`
	Recipes      []Recipe  `gorm:"foreignKey:UploadedByID"`
	Orders       []Order   `gorm:"foreignKey:UserID"`
	Ratings      []Rating  `gorm:"foreignKey:UserID"`
	CartItems    []CartItem `gorm:"foreignKey:UserID"`
}

type Group struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"size:150;uniqueIndex"`
	Users []User `gorm:"many2many:user_groups;"`
}

type Tag struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:50;uniqueIndex;not null"`
	Slug string `gorm:"size:60;uniqueIndex"`
}

func (t *Tag) BeforeSave(tx *gorm.DB) error {
	if t.Slug == "" {
		t.Slug = slugify(t.Name)
	}
	return nil
}

type Recipe struct {
	ID               uint      `gorm:"primaryKey"`
	Title            string    `gorm:"size:255;uniqueIndex;not null"`
	Slug             string    `gorm:"size:255;uniqueIndex"`
	Description      string
	Ingredients      string
	Instructions     string
	Servings         int       `gorm:"default:1"`
	Price            *float64  `gorm:"type:decimal(6,2)"`
	UploadedByID     *uint
	UploadedBy       *User     `gorm:"foreignKey:UploadedByID"`
	Image            string
	ImageThumbnail   string
	ImageSmall       string
	ImageMedium      string
	ImageLarge       string
	IsHighlighted    bool `gorm:"default:false"`
	IsApproved       bool `gorm:"default:false"`
	MaxDailyOrders   *int
	DailyOrdersCount int       `gorm:"default:0"`
	PrepTime         *int
	CookTime         *int
	Calories         *int
	Tags             []Tag     `gorm:"many2many:recipe_tags;"`
	FavoritedBy      []User    `gorm:"many2many:user_favorites;"`
	Ratings          []Rating  `gorm:"foreignKey:RecipeID"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (r *Recipe) BeforeSave(tx *gorm.DB) error {
	if r.Slug == "" {
		r.Slug = slugify(r.Title)
	}
	return nil
}

type Rating struct {
	ID        uint `gorm:"primaryKey"`
	RecipeID  uint `gorm:"uniqueIndex:idx_recipe_user"`
	Recipe    Recipe
	UserID    uint `gorm:"uniqueIndex:idx_recipe_user"`
	User      User
	Score     float64 `gorm:"type:real"`
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Order struct {
	ID            uint `gorm:"primaryKey"`
	UserID        uint
	User          User
	Status        string    `gorm:"size:20;default:PENDING"`
	Completed     bool      `gorm:"default:false"`
	TotalPrice    float64   `gorm:"type:decimal(6,2);default:0"`
	TrackingToken string    `gorm:"type:uuid;uniqueIndex;not null"`
	Items         []OrderItem `gorm:"foreignKey:OrderID"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.TrackingToken == "" {
		o.TrackingToken = uuid.New().String()
	}
	return nil
}

type OrderItem struct {
	ID       uint `gorm:"primaryKey"`
	OrderID  uint
	Order    Order
	RecipeID uint
	Recipe   Recipe
	Quantity int     `gorm:"default:1"`
	Price    float64 `gorm:"type:decimal(6,2)"`
}

type CartItem struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"uniqueIndex:idx_cart_user_recipe"`
	User      User
	RecipeID  uint `gorm:"uniqueIndex:idx_cart_user_recipe"`
	Recipe    Recipe
	Quantity  int     `gorm:"default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (ci CartItem) TotalPrice() float64 {
	if ci.Recipe.Price != nil {
		return *ci.Recipe.Price * float64(ci.Quantity)
	}
	return 0
}

type RecipeHistory struct {
	ID          uint `gorm:"primaryKey"`
	RecipeID    uint
	ChangedByID *uint
	FieldName   string
	OldValue    string
	NewValue    string
	ChangedAt   time.Time
}

func HashedFilename(original string) (string, error) {
	ext := filepath.Ext(original)
	hash := sha256.Sum256([]byte(original + time.Now().String()))
	return "recipes/" + hex.EncodeToString(hash[:16]) + ext, nil
}

func init() {
	_ = io.Discard
}
