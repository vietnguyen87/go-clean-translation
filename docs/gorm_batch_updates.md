# GORM Batch Updates Guide

This guide covers various approaches to perform batch updates using GORM, with examples and best practices for your translation service.

## Table of Contents

1. [Basic Batch Updates](#basic-batch-updates)
2. [Transaction-based Batch Updates](#transaction-based-batch-updates)
3. [Raw SQL Batch Updates](#raw-sql-batch-updates)
4. [Bulk Updates with CASE WHEN](#bulk-updates-with-case-when)
5. [Batch Upserts](#batch-upserts)
6. [Performance Considerations](#performance-considerations)
7. [Best Practices](#best-practices)

## Basic Batch Updates

### Simple Condition-based Updates

```go
// Update all translations with a specific source language
result := db.Model(&Translation{}).
    Where("source = ?", "en").
    Update("status", "verified")

fmt.Printf("Updated %d records\n", result.RowsAffected)
```

### Multiple Field Updates

```go
// Update multiple fields at once
updates := map[string]interface{}{
    "result_text": "Updated text",
    "status":      "processed",
    "updated_at":  time.Now(),
}

result := db.Model(&Translation{}).
    Where("source = ? AND destination = ?", "en", "es").
    Updates(updates)
```

## Transaction-based Batch Updates

### Individual Updates in Transaction

```go
err := db.Transaction(func(tx *gorm.DB) error {
    translations := []Translation{
        {OriginalText: "Hello", Source: "en", Destination: "es", ResultText: "Hola"},
        {OriginalText: "World", Source: "en", Destination: "es", ResultText: "Mundo"},
    }
    
    for _, t := range translations {
        if err := tx.Model(&Translation{}).
            Where("original_text = ? AND source = ? AND destination = ?", 
                t.OriginalText, t.Source, t.Destination).
            Update("result_text", t.ResultText).Error; err != nil {
            return err
        }
    }
    return nil
})
```

**Pros:**
- Atomic operations
- Easy to handle errors
- Flexible logic

**Cons:**
- Multiple SQL queries
- Slower for large batches

## Raw SQL Batch Updates

### Using CASE WHEN for Complex Logic

```go
query := `
    UPDATE translations 
    SET result_text = CASE 
        WHEN source = 'en' AND destination = 'es' THEN 'Spanish translation'
        WHEN source = 'en' AND destination = 'fr' THEN 'French translation'
        ELSE result_text 
    END,
    status = CASE 
        WHEN source = 'en' THEN 'processed'
        ELSE status 
    END
    WHERE id IN (?, ?, ?)
`

result := db.Exec(query, 1, 2, 3)
```

### Parameterized Bulk Updates

```go
// Build dynamic query based on data
var caseStatement string
var args []interface{}

for i, update := range updates {
    if i > 0 {
        caseStatement += " "
    }
    caseStatement += "WHEN (original_text = ? AND source = ? AND destination = ?) THEN ?"
    args = append(args, update.OriginalText, update.Source, update.Destination, update.ResultText)
}

query := "UPDATE translations SET result_text = CASE " + caseStatement + " ELSE result_text END"
result := db.Exec(query, args...)
```

## Bulk Updates with CASE WHEN

### MySQL-specific Optimization

```go
func BatchUpdateResultTextsBulk(db *gorm.DB, updates []Translation) error {
    if len(updates) == 0 {
        return nil
    }

    var caseStatement string
    var whereConditions []string
    var args []interface{}

    for i, update := range updates {
        if i > 0 {
            caseStatement += " "
        }
        caseStatement += "WHEN (original_text = ? AND source = ? AND destination = ?) THEN ?"
        args = append(args, update.OriginalText, update.Source, update.Destination, update.ResultText)
        
        whereConditions = append(whereConditions, "(?, ?, ?)")
        args = append(args, update.OriginalText, update.Source, update.Destination)
    }

    query := "UPDATE translations SET result_text = CASE " + caseStatement + " ELSE result_text END " +
        "WHERE (original_text, source, destination) IN (" + whereConditions[0]
    
    for i := 1; i < len(whereConditions); i++ {
        query += ", " + whereConditions[i]
    }
    query += ")"

    return db.Exec(query, args...).Error
}
```

## Batch Upserts

### Insert or Update Logic

```go
func BatchUpsert(db *gorm.DB, translations []Translation) error {
    return db.Transaction(func(tx *gorm.DB) error {
        for _, translation := range translations {
            var existing Translation
            err := tx.Where("original_text = ? AND source = ? AND destination = ?", 
                translation.OriginalText, translation.Source, translation.Destination).
                First(&existing).Error

            if err == gorm.ErrRecordNotFound {
                // Insert new record
                if err := tx.Create(&translation).Error; err != nil {
                    return err
                }
            } else if err != nil {
                return err
            } else {
                // Update existing record
                if err := tx.Model(&existing).Updates(map[string]interface{}{
                    "result_text": translation.ResultText,
                }).Error; err != nil {
                    return err
                }
            }
        }
        return nil
    })
}
```

## Performance Considerations

### Batch Size Optimization

```go
const BATCH_SIZE = 1000

func BatchUpdateInChunks(db *gorm.DB, updates []Translation) error {
    for i := 0; i < len(updates); i += BATCH_SIZE {
        end := i + BATCH_SIZE
        if end > len(updates) {
            end = len(updates)
        }
        
        chunk := updates[i:end]
        if err := processChunk(db, chunk); err != nil {
            return err
        }
    }
    return nil
}
```

### Index Considerations

Ensure you have proper indexes for your WHERE clauses:

```sql
-- For translation lookups
CREATE INDEX idx_translations_lookup ON translations(original_text, source, destination);

-- For status updates
CREATE INDEX idx_translations_status ON translations(status);

-- For source/destination filtering
CREATE INDEX idx_translations_lang ON translations(source, destination);
```

## Best Practices

### 1. Use Transactions for Related Updates

```go
err := db.Transaction(func(tx *gorm.DB) error {
    // All updates succeed or fail together
    if err := updateTranslations(tx); err != nil {
        return err
    }
    if err := updateMetadata(tx); err != nil {
        return err
    }
    return nil
})
```

### 2. Handle Empty Batches

```go
func BatchUpdate(updates []Translation) error {
    if len(updates) == 0 {
        return nil // Early return for empty batches
    }
    // Process updates...
}
```

### 3. Use Context for Cancellation

```go
func BatchUpdateWithContext(ctx context.Context, updates []Translation) error {
    return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // Context can be cancelled during long-running operations
        for _, update := range updates {
            select {
            case <-ctx.Done():
                return ctx.Err()
            default:
                // Process update
            }
        }
        return nil
    })
}
```

### 4. Error Handling and Logging

```go
func BatchUpdateWithLogging(updates []Translation) error {
    start := time.Now()
    processed := 0
    
    err := db.Transaction(func(tx *gorm.DB) error {
        for _, update := range updates {
            if err := processUpdate(tx, update); err != nil {
                log.Printf("Failed to process update: %v", err)
                return err
            }
            processed++
        }
        return nil
    })
    
    duration := time.Since(start)
    log.Printf("Processed %d updates in %v", processed, duration)
    
    return err
}
```

### 5. Choose the Right Approach

| Approach | Use When | Pros | Cons |
|----------|----------|------|------|
| **Simple Updates** | Single field, simple conditions | Fast, simple | Limited flexibility |
| **Transaction Updates** | Complex logic, error handling | Atomic, flexible | Multiple queries, slower |
| **Raw SQL** | Complex conditions, performance | Single query, fast | Less portable, harder to maintain |
| **CASE WHEN** | Multiple conditional updates | Single query, readable | MySQL-specific, complex queries |

## Example Usage in Your Translation Service

```go
// Update multiple translation results
func (repo mysqlRepo) UpdateTranslationResults(ctx context.Context, updates []entity.Translation) error {
    if len(updates) < 100 {
        // Use transaction approach for small batches
        return repo.BatchUpdateResultTexts(ctx, updates)
    } else {
        // Use bulk approach for large batches
        return repo.BatchUpdateResultTextsBulk(ctx, updates)
    }
}
```

## Testing Batch Updates

```go
func TestBatchUpdate(t *testing.T) {
    db := setupTestDB()
    
    // Create test data
    translations := createTestTranslations(100)
    
    // Test batch update
    err := BatchUpdateResultTexts(db, translations)
    assert.NoError(t, err)
    
    // Verify updates
    for _, t := range translations {
        var result Translation
        db.Where("original_text = ?", t.OriginalText).First(&result)
        assert.Equal(t, t.ResultText, result.ResultText)
    }
}
```

This guide provides comprehensive coverage of GORM batch update patterns. Choose the approach that best fits your specific use case, considering factors like batch size, complexity, and performance requirements.