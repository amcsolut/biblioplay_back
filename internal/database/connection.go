package database

import (
	"fmt"
	"log"
	"strings"

	"api-backend-infinitrum/internal/models/catalog"
	"api-backend-infinitrum/internal/models/feed"
	"api-backend-infinitrum/internal/models/profile"
	"api-backend-infinitrum/internal/models/user"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Initialize(databaseURL string) (*gorm.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	var db *gorm.DB
	var err error

	// Detect database type from URL
	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		// PostgreSQL connection
		db, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}
		log.Println("Connected to PostgreSQL database")
	} else if strings.HasPrefix(databaseURL, "sqlserver://") || strings.Contains(databaseURL, "sqlserver") {
		// SQL Server connection
		db, err = gorm.Open(sqlserver.Open(databaseURL), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQL Server: %w", err)
		}
		log.Println("Connected to SQL Server database")
	} else {
		// Try PostgreSQL first (default), then SQL Server
		db, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			// Try SQL Server if PostgreSQL fails
			log.Printf("PostgreSQL connection failed, trying SQL Server: %v", err)
			db, err = gorm.Open(sqlserver.Open(databaseURL), &gorm.Config{
				Logger: logger.Default.LogMode(logger.Info),
			})
			if err != nil {
				return nil, fmt.Errorf("failed to connect to database (tried PostgreSQL and SQL Server): %w", err)
			}
			log.Println("Connected to SQL Server database")
		} else {
			log.Println("Connected to PostgreSQL database")
		}
	}

	log.Println("Database connection established successfully")
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return AutoMigrateWithURL(db, "")
}

// AutoMigrateWithURL performs database migrations with database URL for better driver detection
func AutoMigrateWithURL(db *gorm.DB, databaseURL string) error {
	migrator := db.Migrator()

	log.Println("Starting database migration...")

	// Migrate User model - this will create table and add new columns
	if err := migrator.AutoMigrate(&user.User{}); err != nil {
		return fmt.Errorf("failed to migrate User: %w", err)
	}

	// Ensure User table has all required columns and indexes
	if migrator.HasTable(&user.User{}) {
		log.Println("Ensuring User table has all required columns and indexes...")
		
		// Add/update columns if they don't exist
		if !migrator.HasColumn(&user.User{}, "provider") {
			log.Println("Adding provider column...")
			if err := migrator.AddColumn(&user.User{}, "provider"); err != nil {
				log.Printf("Warning: Could not add provider column: %v", err)
			} else {
				log.Println("Provider column added successfully")
			}
		}
		
		if !migrator.HasColumn(&user.User{}, "provider_id") {
			log.Println("Adding provider_id column...")
			if err := migrator.AddColumn(&user.User{}, "provider_id"); err != nil {
				log.Printf("Warning: Could not add provider_id column: %v", err)
			} else {
				log.Println("Provider_id column added successfully")
			}
		}
		
		// Ensure indexes exist
		if !migrator.HasIndex(&user.User{}, "provider") {
			log.Println("Creating provider index...")
			if err := migrator.CreateIndex(&user.User{}, "provider"); err != nil {
				log.Printf("Warning: Could not create provider index: %v", err)
			} else {
				log.Println("Provider index created successfully")
			}
		}
		
		if !migrator.HasIndex(&user.User{}, "provider_id") {
			log.Println("Creating provider_id index...")
			if err := migrator.CreateIndex(&user.User{}, "provider_id"); err != nil {
				log.Printf("Warning: Could not create provider_id index: %v", err)
			} else {
				log.Println("Provider_id index created successfully")
			}
		}
		
		// Make password_hash nullable if it's not already
		// Note: This requires raw SQL for some databases
		if err := makePasswordHashNullable(db, databaseURL); err != nil {
			log.Printf("Warning: Could not make password_hash nullable: %v", err)
		}
	}

	// Migrate UserSession model
	if err := migrator.AutoMigrate(&user.UserSession{}); err != nil {
		return fmt.Errorf("failed to migrate UserSession: %w", err)
	}

	// Migrate UserInvitation model
	if err := migrator.AutoMigrate(&user.UserInvitation{}); err != nil {
		return fmt.Errorf("failed to migrate UserInvitation: %w", err)
	}

	// Migrate PasswordReset model
	if err := migrator.AutoMigrate(&user.PasswordReset{}); err != nil {
		return fmt.Errorf("failed to migrate PasswordReset: %w", err)
	}

	// Profile & community (FK users)
	if err := migrator.AutoMigrate(&profile.ProfileMember{}); err != nil {
		return fmt.Errorf("failed to migrate ProfileMember: %w", err)
	}
	if err := migrator.AutoMigrate(&profile.ProfileAuthor{}); err != nil {
		return fmt.Errorf("failed to migrate ProfileAuthor: %w", err)
	}
	if err := migrator.AutoMigrate(&profile.Community{}); err != nil {
		return fmt.Errorf("failed to migrate Community: %w", err)
	}

	// Catálogo (obras + capítulos ebook / audiobook)
	if err := migrator.AutoMigrate(&catalog.CatalogBook{}); err != nil {
		return fmt.Errorf("failed to migrate CatalogBook: %w", err)
	}
	if err := migrator.AutoMigrate(&catalog.EbookChapter{}); err != nil {
		return fmt.Errorf("failed to migrate EbookChapter: %w", err)
	}
	if err := migrator.AutoMigrate(&catalog.AudiobookChapter{}); err != nil {
		return fmt.Errorf("failed to migrate AudiobookChapter: %w", err)
	}

	// Feed da comunidade (posts, mídia, comentários, replies, reações)
	if err := migrator.AutoMigrate(&feed.CommunityPost{}); err != nil {
		return fmt.Errorf("failed to migrate CommunityPost: %w", err)
	}
	if err := migrator.AutoMigrate(&feed.PostMedia{}); err != nil {
		return fmt.Errorf("failed to migrate PostMedia: %w", err)
	}
	if err := migrator.AutoMigrate(&feed.PostComment{}); err != nil {
		return fmt.Errorf("failed to migrate PostComment: %w", err)
	}
	if err := migrator.AutoMigrate(&feed.CommentReply{}); err != nil {
		return fmt.Errorf("failed to migrate CommentReply: %w", err)
	}
	if err := migrator.AutoMigrate(&feed.Reaction{}); err != nil {
		return fmt.Errorf("failed to migrate Reaction: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

// makePasswordHashNullable makes the password_hash column nullable
// This is needed because GORM AutoMigrate doesn't change column nullability by default
func makePasswordHashNullable(db *gorm.DB, databaseURL string) error {
	// Detect database type from URL
	var alterSQL string
	var checkSQL string
	
	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		// PostgreSQL
		alterSQL = "ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL"
		checkSQL = "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'password_hash' AND is_nullable = 'NO'"
	} else if strings.HasPrefix(databaseURL, "sqlserver://") || strings.Contains(databaseURL, "sqlserver") {
		// SQL Server
		alterSQL = "ALTER TABLE users ALTER COLUMN password_hash VARCHAR(255) NULL"
		checkSQL = "SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'users' AND COLUMN_NAME = 'password_hash' AND IS_NULLABLE = 'NO'"
	} else if databaseURL != "" {
		// Try PostgreSQL syntax as default
		alterSQL = "ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL"
		checkSQL = "SELECT COUNT(*) FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'password_hash' AND is_nullable = 'NO'"
	} else {
		// No URL provided, skip
		return nil
	}

	// Check if column exists and is NOT NULL before altering
	var count int
	if err := db.Raw(checkSQL).Scan(&count).Error; err != nil {
		// If check fails (might be different database or column doesn't exist), skip
		log.Printf("Could not check password_hash nullability: %v", err)
		return nil
	}
	
	if count > 0 {
		// Column exists and is NOT NULL, make it nullable
		log.Println("Making password_hash column nullable...")
		if err := db.Exec(alterSQL).Error; err != nil {
			// Log but don't fail - column might already be nullable or syntax might differ
			log.Printf("Could not make password_hash nullable (this is OK if it's already nullable): %v", err)
			return nil
		}
		log.Println("Made password_hash column nullable successfully")
	} else {
		log.Println("password_hash column is already nullable or doesn't exist")
	}

	return nil
}

func TestConnection(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}
