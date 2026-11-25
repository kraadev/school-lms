package helpcenter

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"school-lms/internal/models"
)

type Service struct{ DB *gorm.DB }

func NewService(db *gorm.DB) *Service { return &Service{DB: db} }

type FAQInput struct {
	Question  string `json:"question" binding:"required"`
	Answer    string `json:"answer" binding:"required"`
	Category  string `json:"category"`
	Published *bool  `json:"published"`
}

func (s *Service) List(role string) ([]models.FAQ, error) {
	query := s.DB.Preload("Creator").Order("category ASC, created_at DESC")
	if role != "admin" {
		query = query.Where("published = ?", true)
	}
	var items []models.FAQ
	return items, query.Find(&items).Error
}

func (s *Service) Create(userID string, input FAQInput) (*models.FAQ, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("user id tidak valid")
	}
	category := strings.TrimSpace(input.Category)
	if category == "" {
		category = "Umum"
	}
	published := false
	if input.Published != nil {
		published = *input.Published
	}
	item := &models.FAQ{Question: strings.TrimSpace(input.Question), Answer: strings.TrimSpace(input.Answer), Category: category, Published: published, CreatedBy: &uid}
	if item.Question == "" || item.Answer == "" {
		return nil, errors.New("pertanyaan dan jawaban wajib diisi")
	}
	return item, s.DB.Create(item).Error
}

func (s *Service) Update(id string, input FAQInput) (*models.FAQ, error) {
	var item models.FAQ
	if err := s.DB.First(&item, "id = ?", id).Error; err != nil {
		return nil, err
	}
	category := strings.TrimSpace(input.Category)
	if category == "" {
		category = "Umum"
	}
	updates := map[string]any{"question": strings.TrimSpace(input.Question), "answer": strings.TrimSpace(input.Answer), "category": category}
	if input.Published != nil {
		updates["published"] = *input.Published
	}
	if err := s.DB.Model(&item).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &item, s.DB.First(&item, "id = ?", item.ID).Error
}

func (s *Service) Delete(id string) error { return s.DB.Delete(&models.FAQ{}, "id = ?", id).Error }
