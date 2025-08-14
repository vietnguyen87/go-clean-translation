package main

import (
	"context"
	"fmt"
	"log"

	"go-clean-translation/infras/mysql"
	"go-clean-translation/service"
	"go-clean-translation/service/entity"
)

func main() {
	// This is a demonstration of how to use batch updates in your translation service
	// In a real application, you would initialize your database connection and services

	ctx := context.Background()

	// Example 1: Batch update translation results
	exampleBatchUpdateTranslations(ctx)

	// Example 2: Batch upsert translations
	exampleBatchUpsertTranslations(ctx)

	// Example 3: Conditional batch updates
	exampleConditionalBatchUpdates(ctx)
}

func exampleBatchUpdateTranslations(ctx context.Context) {
	fmt.Println("\n=== Example 1: Batch Update Translation Results ===")

	// Create sample translation updates
	updates := []entity.Translation{
		entity.NewTranslation("Hello", "en", "es", "Hola"),
		entity.NewTranslation("World", "en", "es", "Mundo"),
		entity.NewTranslation("Good morning", "en", "es", "Buenos días"),
		entity.NewTranslation("Thank you", "en", "es", "Gracias"),
	}

	// In a real application, you would get these from your service layer
	// service := service.NewService(repository, googleService)
	// err := service.BatchUpdateTranslations(ctx, updates)

	fmt.Printf("Would update %d translations:\n", len(updates))
	for _, t := range updates {
		fmt.Printf("  %s (%s→%s): %s\n", t.OriginalText, t.Source, t.Destination, t.ResultText)
	}
}

func exampleBatchUpsertTranslations(ctx context.Context) {
	fmt.Println("\n=== Example 2: Batch Upsert Translations ===")

	// Create translations that might be new or existing
	translations := []entity.Translation{
		entity.NewTranslation("Hello", "en", "fr", "Bonjour"),
		entity.NewTranslation("Goodbye", "en", "fr", "Au revoir"),
		entity.NewTranslation("Please", "en", "fr", "S'il vous plaît"),
		entity.NewTranslation("Excuse me", "en", "fr", "Excusez-moi"),
	}

	// In a real application:
	// err := service.BatchUpsertTranslations(ctx, translations)

	fmt.Printf("Would upsert %d translations:\n", len(translations))
	for _, t := range translations {
		fmt.Printf("  %s (%s→%s): %s\n", t.OriginalText, t.Source, t.Destination, t.ResultText)
	}
}

func exampleConditionalBatchUpdates(ctx context.Context) {
	fmt.Println("\n=== Example 3: Conditional Batch Updates ===")

	// Example: Update all English to Spanish translations to have a specific status
	// This would use the BatchUpdateByCondition method

	condition := map[string]interface{}{
		"source":        "en",
		"destination":   "es",
	}

	updates := map[string]interface{}{
		"status": "verified",
		// You could add more fields to update
	}

	fmt.Printf("Would update translations matching condition: %+v\n", condition)
	fmt.Printf("With updates: %+v\n", updates)

	// In a real application:
	// err := repository.BatchUpdateByCondition(ctx, condition, updates)
}

// Example of how to use batch updates in a real application
func demonstrateRealUsage() {
	// This shows how you would integrate batch updates into your existing service

	// 1. Initialize your services (this would be done in your main function or dependency injection)
	// db := initializeDatabase()
	// repository := mysql.NewMySQLRepo(db)
	// googleService := google.NewGoogleService()
	// service := service.NewService(repository, googleService)

	// 2. Use batch updates in your application logic
	ctx := context.Background()

	// Example: Process multiple translation requests
	translationRequests := []struct {
		originalText string
		source       string
		destination  string
	}{
		{"Hello", "en", "es"},
		{"World", "en", "es"},
		{"Good morning", "en", "es"},
	}

	// Process translations and collect results
	var updates []entity.Translation
	for _, req := range translationRequests {
		// In a real app, you might call Google Translate API here
		translation := entity.NewTranslation(
			req.originalText,
			req.source,
			req.destination,
			"Translated text", // This would be the actual translation
		)
		updates = append(updates, translation)
	}

	// Batch update all translations
	// if err := service.BatchUpdateTranslations(ctx, updates); err != nil {
	//     log.Printf("Failed to batch update translations: %v", err)
	//     return
	// }

	fmt.Printf("Successfully processed %d translation requests\n", len(updates))
}

// Example of error handling and logging for batch operations
func demonstrateErrorHandling() {
	ctx := context.Background()

	// Create a large batch of updates
	var updates []entity.Translation
	for i := 0; i < 1000; i++ {
		translation := entity.NewTranslation(
			fmt.Sprintf("Text %d", i),
			"en",
			"es",
			fmt.Sprintf("Translation %d", i),
		)
		updates = append(updates, translation)
	}

	// Process in chunks for better error handling
	const chunkSize = 100
	var processed int

	for i := 0; i < len(updates); i += chunkSize {
		end := i + chunkSize
		if end > len(updates) {
			end = len(updates)
		}

		chunk := updates[i:end]
		
		// In a real application:
		// if err := service.BatchUpdateTranslations(ctx, chunk); err != nil {
		//     log.Printf("Failed to process chunk %d-%d: %v", i, end-1, err)
		//     // You might want to continue with other chunks or return early
		//     continue
		// }

		processed += len(chunk)
		log.Printf("Processed chunk %d-%d (%d total processed)", i, end-1, processed)
	}

	log.Printf("Completed processing %d translations in chunks of %d", len(updates), chunkSize)
}