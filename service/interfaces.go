package service

import (
	"context"
	"go-clean-translation/service/entity"
)

type TranslateUseCase interface {
	Translate(ctx context.Context, orgText, source, dest string) (entity.Translation, error)
	FetchHistories(ctx context.Context) ([]entity.Translation, error)
	
	// Batch update methods
	BatchUpdateTranslations(ctx context.Context, updates []entity.Translation) error
	BatchUpsertTranslations(ctx context.Context, translations []entity.Translation) error
}

type TranslateRepository interface {
	GetTranslation(ctx context.Context, orgText, source, dest string) (entity.Translation, error)
	FindHistories(ctx context.Context) ([]entity.Translation, error)
	InsertTranslation(ctx context.Context, translation entity.Translation) error
	
	// Batch update methods
	BatchUpdateResultTexts(ctx context.Context, updates []entity.Translation) error
	BatchUpdateResultTextsBulk(ctx context.Context, updates []entity.Translation) error
	BatchUpdateByCondition(ctx context.Context, condition map[string]interface{}, updates map[string]interface{}) error
	BatchUpsert(ctx context.Context, translations []entity.Translation) error
}

type GoogleService interface {
	Translate(ctx context.Context, orgText, source, dest string) (entity.Translation, error)
}
