package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func MigrateFromDjango(djangoDBPath string) error {
	if _, err := os.Stat(djangoDBPath); os.IsNotExist(err) {
		return nil
	}

	var gormCount int64
	DB.Model(&User{}).Count(&gormCount)
	if gormCount > 0 {
		return nil
	}

	dj, err := sql.Open("sqlite3", djangoDBPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("open django db: %w", err)
	}
	defer dj.Close()

	var migrationCount int
	dj.QueryRow("SELECT COUNT(*) FROM django_migrations").Scan(&migrationCount)
	if migrationCount == 0 {
		return nil
	}

	log.Println("=== Migrating data from Django database ===")

	if err := migrateUsers(dj); err != nil {
		return fmt.Errorf("users: %w", err)
	}
	if err := migrateGroups(dj); err != nil {
		return fmt.Errorf("groups: %w", err)
	}
	if err := migrateTags(dj); err != nil {
		return fmt.Errorf("tags: %w", err)
	}
	if err := migrateSettings(dj); err != nil {
		return fmt.Errorf("settings: %w", err)
	}
	if err := migrateRecipes(dj); err != nil {
		return fmt.Errorf("recipes: %w", err)
	}
	if err := migrateRatings(dj); err != nil {
		return fmt.Errorf("ratings: %w", err)
	}
	if err := migrateOrders(dj); err != nil {
		return fmt.Errorf("orders: %w", err)
	}
	if err := migrateCartItems(dj); err != nil {
		return fmt.Errorf("cartitems: %w", err)
	}
	if err := migrateM2M(dj); err != nil {
		return fmt.Errorf("m2m: %w", err)
	}

	log.Println("=== Django migration complete ===")
	return nil
}

func migrateUsers(dj *sql.DB) error {
	rows, err := dj.Query(`SELECT id, password, last_login, is_superuser, username,
		first_name, last_name, email, is_staff, is_active, date_joined,
		avatar, bio, language, theme
		FROM sandwitches_user ORDER BY id`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var u User
		var avatar, language, theme sql.NullString
		err := rows.Scan(&u.ID, &u.Password, &u.LastLogin, &u.IsSuperuser, &u.Username,
			&u.FirstName, &u.LastName, &u.Email, &u.IsStaff, &u.IsActive, &u.DateJoined,
			&avatar, &u.Bio, &language, &theme)
		if err != nil {
			return fmt.Errorf("scan user: %w", err)
		}
		if language.Valid {
			u.Language = language.String
		} else {
			u.Language = "en"
		}
		if theme.Valid {
			u.Theme = theme.String
		} else {
			u.Theme = "light"
		}
		if avatar.Valid && avatar.String != "" {
			u.Avatar = "/media/" + avatar.String
		}

		// Django uses PBKDF2 hashes, Go uses bcrypt — replace with bcrypt hash of "admin"
		if strings.HasPrefix(u.Password, "pbkdf2_sha256$") || strings.HasPrefix(u.Password, "pbkdf2_") {
			hashed, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
			if err == nil {
				u.Password = string(hashed)
				log.Printf("  User %s: converted Django password to bcrypt (login with 'admin')", u.Username)
			}
		}

		DB.Create(&u)
		count++
	}
	log.Printf("  Migrated %d users", count)
	return rows.Err()
}

func migrateGroups(dj *sql.DB) error {
	rows, err := dj.Query(`SELECT id, name FROM auth_group ORDER BY id`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Name); err != nil {
			return err
		}
		var existing Group
		if DB.Where("id = ?", g.ID).First(&existing).Error == nil {
			continue
		}
		DB.Exec("INSERT INTO groups (id, name) VALUES (?, ?)", g.ID, g.Name)
		count++
	}

	m2mRows, _ := dj.Query(`SELECT id, user_id, group_id FROM sandwitches_user_groups`)
	if m2mRows != nil {
		defer m2mRows.Close()
		for m2mRows.Next() {
			var id, userID, groupID int
			m2mRows.Scan(&id, &userID, &groupID)
			DB.Exec("INSERT INTO user_groups (user_id, group_id) VALUES (?, ?)", userID, groupID)
		}
	}

	log.Printf("  Migrated %d groups", count)
	return rows.Err()
}

func migrateTags(dj *sql.DB) error {
	rows, err := dj.Query(`SELECT id, name, slug FROM sandwitches_tag ORDER BY id`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return err
		}
		DB.Create(&t)
		count++
	}
	log.Printf("  Migrated %d tags", count)
	return rows.Err()
}

func migrateSettings(dj *sql.DB) error {
	var count int
	dj.QueryRow("SELECT COUNT(*) FROM sandwitches_setting").Scan(&count)
	if count == 0 {
		return nil
	}

	var s Setting
	err := dj.QueryRow(`SELECT id, site_name, COALESCE(site_description,''), COALESCE(email,''),
		COALESCE(ai_connection_point,''), COALESCE(ai_model,''), COALESCE(ai_api_key,''),
		COALESCE(gotify_token,''), COALESCE(gotify_url,''), COALESCE(log_level,'INFO')
		FROM sandwitches_setting LIMIT 1`).Scan(
		&s.ID, &s.SiteName, &s.SiteDescription, &s.Email,
		&s.AIConnectionPoint, &s.AIModel, &s.AIAPIKey,
		&s.GotifyToken, &s.GotifyURL, &s.LogLevel)
	if err != nil {
		return err
	}
	DB.Exec("DELETE FROM settings")
	DB.Create(&s)
	log.Println("  Migrated settings")
	return nil
}

func migrateRecipes(dj *sql.DB) error {
	rows, err := dj.Query(`SELECT id, title, slug, description, ingredients, instructions,
		COALESCE(image,''), uploaded_by_id, servings, is_highlighted, price,
		daily_orders_count, max_daily_orders, is_approved,
		calories, cook_time, prep_time, created_at, updated_at
		FROM sandwitches_recipe ORDER BY id`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var r Recipe
		var image string
		var uploadedByID sql.NullInt64
		var price sql.NullFloat64
		var maxDaily *int
		var calories, cookTime, prepTime *int

		err := rows.Scan(&r.ID, &r.Title, &r.Slug, &r.Description, &r.Ingredients, &r.Instructions,
			&image, &uploadedByID, &r.Servings, &r.IsHighlighted, &price,
			&r.DailyOrdersCount, &maxDaily, &r.IsApproved,
			&calories, &cookTime, &prepTime, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return fmt.Errorf("scan recipe: %w", err)
		}

		if image != "" {
			r.Image = "/media/" + image
			r.ImageThumbnail = r.Image
			r.ImageSmall = r.Image
			r.ImageMedium = r.Image
			r.ImageLarge = r.Image
		}
		if uploadedByID.Valid {
			uid := uint(uploadedByID.Int64)
			r.UploadedByID = &uid
		}
		if price.Valid {
			r.Price = &price.Float64
		}
		r.MaxDailyOrders = maxDaily
		r.Calories = calories
		r.CookTime = cookTime
		r.PrepTime = prepTime

		DB.Create(&r)
		count++
	}
	log.Printf("  Migrated %d recipes", count)
	return rows.Err()
}

func migrateRatings(dj *sql.DB) error {
	rows, err := dj.Query(`SELECT id, recipe_id, user_id, score, COALESCE(comment,''),
		created_at, updated_at FROM sandwitches_rating ORDER BY id`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var r Rating
		if err := rows.Scan(&r.ID, &r.RecipeID, &r.UserID, &r.Score, &r.Comment,
			&r.CreatedAt, &r.UpdatedAt); err != nil {
			return err
		}
		DB.Create(&r)
		count++
	}
	log.Printf("  Migrated %d ratings", count)
	return rows.Err()
}

func migrateOrders(dj *sql.DB) error {
	rows, err := dj.Query(`SELECT id, user_id, status, completed, total_price, tracking_token,
		created_at, updated_at FROM sandwitches_order ORDER BY id`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var orderIDs []uint
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.Completed, &o.TotalPrice,
			&o.TrackingToken, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return err
		}
		DB.Create(&o)
		orderIDs = append(orderIDs, o.ID)
	}
	log.Printf("  Migrated %d orders", len(orderIDs))

	itemRows, err := dj.Query(`SELECT id, order_id, recipe_id, quantity, price
		FROM sandwitches_orderitem ORDER BY id`)
	if err != nil {
		return err
	}
	defer itemRows.Close()

	var itemCount int
	for itemRows.Next() {
		var oi OrderItem
		if err := itemRows.Scan(&oi.ID, &oi.OrderID, &oi.RecipeID, &oi.Quantity, &oi.Price); err != nil {
			return err
		}
		DB.Create(&oi)
		itemCount++
	}
	log.Printf("  Migrated %d order items", itemCount)
	return rows.Err()
}

func migrateCartItems(dj *sql.DB) error {
	rows, err := dj.Query(`SELECT id, user_id, recipe_id, quantity, created_at, updated_at
		FROM sandwitches_cartitem ORDER BY id`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var ci CartItem
		if err := rows.Scan(&ci.ID, &ci.UserID, &ci.RecipeID, &ci.Quantity,
			&ci.CreatedAt, &ci.UpdatedAt); err != nil {
			return err
		}
		DB.Create(&ci)
		count++
	}
	log.Printf("  Migrated %d cart items", count)
	return rows.Err()
}

func migrateM2M(dj *sql.DB) error {
	rows, err := dj.Query(`SELECT id, recipe_id, tag_id FROM sandwitches_recipe_tags`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id, recipeID, tagID int
		rows.Scan(&id, &recipeID, &tagID)
		DB.Exec("INSERT INTO recipe_tags (recipe_id, tag_id) VALUES (?, ?)", recipeID, tagID)
		count++
	}
	log.Printf("  Migrated %d recipe-tag relations", count)

	favRows, err := dj.Query(`SELECT id, user_id, recipe_id FROM sandwitches_user_favorites`)
	if err != nil {
		return err
	}
	defer favRows.Close()

	var favCount int
	for favRows.Next() {
		var id, userID, recipeID int
		favRows.Scan(&id, &userID, &recipeID)
		DB.Exec("INSERT INTO user_favorites (user_id, recipe_id) VALUES (?, ?)", userID, recipeID)
		favCount++
	}
	log.Printf("  Migrated %d user favorites", favCount)
	return nil
}


