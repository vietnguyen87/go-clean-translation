package mysql

import (
	"context"
	"go-clean-translation/service/entity"
	"gorm.io/gorm"
)

// BatchUpdateResultTexts updates multiple translations' result texts in a single query
func (repo mysqlRepo) BatchUpdateResultTexts(ctx context.Context, updates []entity.Translation) error {
	if len(updates) == 0 {
		return nil
	}

	// Method 1: Using Transaction with individual updates (safer for complex logic)
	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, update := range updates {
			if err := tx.Table(tbName).
				Where("original_text = ? AND source = ? AND destination = ?", 
					update.OriginalText, update.Source, update.Destination).
				Update("result_text", update.ResultText).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchUpdateResultTextsBulk updates multiple translations using a single SQL query
// This is more efficient for large batches but less flexible
func (repo mysqlRepo) BatchUpdateResultTextsBulk(ctx context.Context, updates []entity.Translation) error {
	if len(updates) == 0 {
		return nil
	}

	// Method 2: Using CASE WHEN for bulk update (MySQL specific)
	// This generates a single SQL query like:
	// UPDATE translations SET result_text = CASE 
	//   WHEN (original_text = 'text1' AND source = 'en' AND destination = 'es') THEN 'result1'
	//   WHEN (original_text = 'text2' AND source = 'en' AND destination = 'es') THEN 'result2'
	//   ELSE result_text END
	// WHERE (original_text, source, destination) IN (('text1','en','es'), ('text2','en','es'))

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

	query := "UPDATE " + tbName + " SET result_text = CASE " + caseStatement + " ELSE result_text END " +
		"WHERE (original_text, source, destination) IN (" + whereConditions[0]
	
	for i := 1; i < len(whereConditions); i++ {
		query += ", " + whereConditions[i]
	}
	query += ")"

	return repo.db.WithContext(ctx).Exec(query, args...).Error
}

// BatchUpdateByCondition updates multiple records based on a condition
func (repo mysqlRepo) BatchUpdateByCondition(ctx context.Context, condition map[string]interface{}, updates map[string]interface{}) error {
	return repo.db.WithContext(ctx).Table(tbName).Where(condition).Updates(updates).Error
}

// BatchUpdateWithMap updates multiple records using a map of conditions to updates
func (repo mysqlRepo) BatchUpdateWithMap(ctx context.Context, updateMap map[entity.Translation]string) error {
	if len(updateMap) == 0 {
		return nil
	}

	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for translation, newResultText := range updateMap {
			if err := tx.Table(tbName).
				Where("original_text = ? AND source = ? AND destination = ?", 
					translation.OriginalText, translation.Source, translation.Destination).
				Update("result_text", newResultText).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchUpdateStatus updates multiple translations with a status field (if you add one)
// This demonstrates updating multiple fields at once
func (repo mysqlRepo) BatchUpdateStatus(ctx context.Context, translations []entity.Translation, status string) error {
	if len(translations) == 0 {
		return nil
	}

	var ids []string
	for _, t := range translations {
		// Create a unique identifier for each translation
		id := t.OriginalText + "|" + t.Source + "|" + t.Destination
		ids = append(ids, id)
	}

	// Update all matching records in one query
	return repo.db.WithContext(ctx).Table(tbName).
		Where("CONCAT(original_text, '|', source, '|', destination) IN ?", ids).
		Update("status", status).Error
}

// BatchUpsert performs batch insert or update operations
func (repo mysqlRepo) BatchUpsert(ctx context.Context, translations []entity.Translation) error {
	if len(translations) == 0 {
		return nil
	}

	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, translation := range translations {
			// Try to find existing record
			var existing entity.Translation
			err := tx.Table(tbName).
				Where("original_text = ? AND source = ? AND destination = ?", 
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
				if err := tx.Table(tbName).
					Where("original_text = ? AND source = ? AND destination = ?", 
						translation.OriginalText, translation.Source, translation.Destination).
					Updates(map[string]interface{}{
						"result_text": translation.ResultText,
					}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}