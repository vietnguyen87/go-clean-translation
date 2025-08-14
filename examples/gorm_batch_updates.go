package main

import (
	"context"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// User model for demonstration
type User struct {
	ID       uint   `gorm:"primarykey"`
	Name     string `gorm:"size:255"`
	Email    string `gorm:"size:255;uniqueIndex"`
	Status   string `gorm:"size:50;default:'active'"`
	Age      int
	IsActive bool `gorm:"default:true"`
}

// Product model for demonstration
type Product struct {
	ID          uint    `gorm:"primarykey"`
	Name        string  `gorm:"size:255"`
	Price       float64 `gorm:"type:decimal(10,2)"`
	Stock       int     `gorm:"default:0"`
	Category    string  `gorm:"size:100"`
	IsAvailable bool    `gorm:"default:true"`
}

func main() {
	// Connect to database (replace with your connection details)
	dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate tables
	db.AutoMigrate(&User{}, &Product{})

	ctx := context.Background()

	// Example 1: Simple batch update with condition
	example1SimpleBatchUpdate(db, ctx)

	// Example 2: Batch update multiple fields
	example2BatchUpdateMultipleFields(db, ctx)

	// Example 3: Batch update with transaction
	example3BatchUpdateWithTransaction(db, ctx)

	// Example 4: Batch update using raw SQL
	example4BatchUpdateWithRawSQL(db, ctx)

	// Example 5: Batch upsert (insert or update)
	example5BatchUpsert(db, ctx)

	// Example 6: Batch update with map
	example6BatchUpdateWithMap(db, ctx)

	// Example 7: Conditional batch update
	example7ConditionalBatchUpdate(db, ctx)

	// Example 8: Batch update with joins
	example8BatchUpdateWithJoins(db, ctx)
}

// Example 1: Simple batch update with condition
func example1SimpleBatchUpdate(db *gorm.DB, ctx context.Context) {
	fmt.Println("\n=== Example 1: Simple Batch Update ===")

	// Update all users with age > 25 to have status 'senior'
	result := db.WithContext(ctx).Model(&User{}).
		Where("age > ?", 25).
		Update("status", "senior")

	fmt.Printf("Updated %d users to senior status\n", result.RowsAffected)
}

// Example 2: Batch update multiple fields
func example2BatchUpdateMultipleFields(db *gorm.DB, ctx context.Context) {
	fmt.Println("\n=== Example 2: Batch Update Multiple Fields ===")

	// Update multiple fields for all products in a specific category
	updates := map[string]interface{}{
		"price":        gorm.Expr("price * ?", 1.1), // Increase price by 10%
		"is_available": true,
		"stock":        gorm.Expr("stock + ?", 100), // Add 100 to stock
	}

	result := db.WithContext(ctx).Model(&Product{}).
		Where("category = ?", "electronics").
		Updates(updates)

	fmt.Printf("Updated %d electronics products\n", result.RowsAffected)
}

// Example 3: Batch update with transaction
func example3BatchUpdateWithTransaction(db *gorm.DB, ctx context.Context) {
	fmt.Println("\n=== Example 3: Batch Update with Transaction ===")

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update user statuses in batches
		userIDs := []uint{1, 2, 3, 4, 5}
		
		for _, id := range userIDs {
			if err := tx.Model(&User{}).
				Where("id = ?", id).
				Update("status", "verified").Error; err != nil {
				return err
			}
		}

		// Update product prices
		if err := tx.Model(&Product{}).
			Where("price < ?", 50.0).
			Update("price", gorm.Expr("price * ?", 1.05)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Transaction failed: %v\n", err)
	} else {
		fmt.Println("Transaction completed successfully")
	}
}

// Example 4: Batch update using raw SQL
func example4BatchUpdateWithRawSQL(db *gorm.DB, ctx context.Context) {
	fmt.Println("\n=== Example 4: Batch Update with Raw SQL ===")

	// Using CASE WHEN for complex conditional updates
	query := `
		UPDATE users 
		SET status = CASE 
			WHEN age < 18 THEN 'minor'
			WHEN age BETWEEN 18 AND 25 THEN 'young'
			WHEN age BETWEEN 26 AND 50 THEN 'adult'
			ELSE 'senior'
		END,
		is_active = CASE 
			WHEN age >= 18 THEN true
			ELSE false
		END
		WHERE id IN (?, ?, ?, ?)
	`

	result := db.WithContext(ctx).Exec(query, 1, 2, 3, 4)
	fmt.Printf("Updated %d users with complex logic\n", result.RowsAffected)
}

// Example 5: Batch upsert (insert or update)
func example5BatchUpsert(db *gorm.DB, ctx context.Context) {
	fmt.Println("\n=== Example 5: Batch Upsert ===")

	users := []User{
		{Name: "John Doe", Email: "john@example.com", Age: 30, Status: "active"},
		{Name: "Jane Smith", Email: "jane@example.com", Age: 25, Status: "active"},
		{Name: "Bob Johnson", Email: "bob@example.com", Age: 35, Status: "inactive"},
	}

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, user := range users {
			var existing User
			err := tx.Where("email = ?", user.Email).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				// Insert new user
				if err := tx.Create(&user).Error; err != nil {
					return err
				}
				fmt.Printf("Inserted new user: %s\n", user.Name)
			} else if err != nil {
				return err
			} else {
				// Update existing user
				if err := tx.Model(&existing).Updates(map[string]interface{}{
					"name":   user.Name,
					"age":    user.Age,
					"status": user.Status,
				}).Error; err != nil {
					return err
				}
				fmt.Printf("Updated existing user: %s\n", user.Name)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Upsert failed: %v\n", err)
	}
}

// Example 6: Batch update with map
func example6BatchUpdateWithMap(db *gorm.DB, ctx context.Context) {
	fmt.Println("\n=== Example 6: Batch Update with Map ===")

	// Create a map of user IDs to new statuses
	statusUpdates := map[uint]string{
		1: "premium",
		2: "vip",
		3: "regular",
	}

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for userID, newStatus := range statusUpdates {
			if err := tx.Model(&User{}).
				Where("id = ?", userID).
				Update("status", newStatus).Error; err != nil {
				return err
			}
			fmt.Printf("Updated user %d to status: %s\n", userID, newStatus)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Map update failed: %v\n", err)
	}
}

// Example 7: Conditional batch update
func example7ConditionalBatchUpdate(db *gorm.DB, ctx context.Context) {
	fmt.Println("\n=== Example 7: Conditional Batch Update ===")

	// Update products based on multiple conditions
	result := db.WithContext(ctx).Model(&Product{}).
		Where("category = ? AND price < ? AND stock > ?", "books", 100.0, 0).
		Updates(map[string]interface{}{
			"is_available": true,
			"price":        gorm.Expr("price * ?", 0.9), // 10% discount
		})

	fmt.Printf("Updated %d books with discount\n", result.RowsAffected)
}

// Example 8: Batch update with joins
func example8BatchUpdateWithJoins(db *gorm.DB, ctx context.Context) {
	fmt.Println("\n=== Example 8: Batch Update with Joins ===")

	// This example shows how you might update products based on user preferences
	// (assuming you have a user_preferences table)
	
	// For demonstration, we'll update products that are in low stock
	result := db.WithContext(ctx).Model(&Product{}).
		Joins("JOIN (SELECT category, AVG(stock) as avg_stock FROM products GROUP BY category) as avg_stocks ON products.category = avg_stocks.category").
		Where("products.stock < avg_stocks.avg_stock * 0.5").
		Update("is_available", false)

	fmt.Printf("Updated %d low-stock products to unavailable\n", result.RowsAffected)
}

// Utility function to create sample data
func createSampleData(db *gorm.DB) {
	// Create sample users
	users := []User{
		{Name: "Alice", Email: "alice@example.com", Age: 25, Status: "active"},
		{Name: "Bob", Email: "bob@example.com", Age: 30, Status: "active"},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35, Status: "inactive"},
	}

	for _, user := range users {
		db.Create(&user)
	}

	// Create sample products
	products := []Product{
		{Name: "Laptop", Price: 999.99, Stock: 10, Category: "electronics", IsAvailable: true},
		{Name: "Book", Price: 29.99, Stock: 50, Category: "books", IsAvailable: true},
		{Name: "Phone", Price: 699.99, Stock: 5, Category: "electronics", IsAvailable: true},
	}

	for _, product := range products {
		db.Create(&product)
	}
}