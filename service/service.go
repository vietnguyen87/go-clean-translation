package service

import (
	"context"
	"go-clean-translation/service/entity"
)

type service struct {
	repository    TranslateRepository
	googleService GoogleService
}

func NewService(repository TranslateRepository, googleService GoogleService) service {
	return service{repository: repository, googleService: googleService}
}

func (s service) Translate(ctx context.Context, orgText, source, dest string) (entity.Translation, error) {
	oldTrans, err := s.repository.GetTranslation(ctx, orgText, source, dest)

	if err == nil {
		return oldTrans, nil
	}

	// TODO: should check case other db error.

	newTrans, err := s.googleService.Translate(ctx, orgText, source, dest)

	if err != nil {
		return entity.Translation{}, err
	}

	go func() {
		_ = s.repository.InsertTranslation(ctx, newTrans)
	}()

	return newTrans, nil
}

func (s service) FetchHistories(ctx context.Context) ([]entity.Translation, error) {
	return s.repository.FindHistories(ctx)
}

// BatchUpdateTranslations updates multiple translation results
func (s service) BatchUpdateTranslations(ctx context.Context, updates []entity.Translation) error {
	if len(updates) == 0 {
		return nil
	}

	// Choose the appropriate batch update method based on batch size
	if len(updates) < 100 {
		// Use transaction approach for small batches
		return s.repository.BatchUpdateResultTexts(ctx, updates)
	} else {
		// Use bulk approach for large batches
		return s.repository.BatchUpdateResultTextsBulk(ctx, updates)
	}
}

// BatchUpsertTranslations performs batch insert or update operations
func (s service) BatchUpsertTranslations(ctx context.Context, translations []entity.Translation) error {
	return s.repository.BatchUpsert(ctx, translations)
}
