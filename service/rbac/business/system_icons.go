package services

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sitehub/common/errorx"
	"sitehub/service/rbac/model/entity"
)

type SystemIcons struct{ db *gorm.DB }

func NewSystemIcons(db *gorm.DB) *SystemIcons { return &SystemIcons{db: db} }

func (s *SystemIcons) List(ctx context.Context) ([]entity.SystemIcon, error) {
	var icons []entity.SystemIcon
	err := s.db.WithContext(ctx).Select("id", "name", "label", "category").Order("category, name").Find(&icons).Error
	return icons, err
}

func (s *SystemIcons) Content(ctx context.Context, name string) (string, error) {
	var icon entity.SystemIcon
	err := s.db.WithContext(ctx).Unscoped().Select("svg").Where("name = ?", name).Take(&icon).Error
	return icon.SVG, err
}

func (s *SystemIcons) Validate(input SystemIconInput) (*entity.SystemIcon, error) {
	icon, err := normalizeSystemIconInput(input)
	if err != nil {
		return nil, errorx.Invalid(err.Error())
	}
	return icon, nil
}

func (s *SystemIcons) Create(ctx context.Context, input SystemIconInput) (*entity.SystemIcon, error) {
	icon, err := s.Validate(input)
	if err != nil {
		return nil, err
	}
	err = s.db.WithContext(ctx).Create(icon).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, errorx.Conflicting("图标标识已被使用")
	}
	return icon, err
}

func (s *SystemIcons) Update(ctx context.Context, id uint64, input SystemIconInput) (*entity.SystemIcon, error) {
	if id == 0 {
		return nil, errorx.Invalid("图标 ID 无效")
	}
	icon, err := s.Validate(input)
	if err != nil {
		return nil, err
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current entity.SystemIcon
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, id).Error; err != nil {
			return err
		}
		if icon.Name != current.Name {
			return errorx.Conflicting("已创建的图标标识不可更改")
		}
		icon.ID = current.ID
		return tx.Model(&current).Select("label", "category", "svg").Updates(icon).Error
	})
	return icon, err
}

func (s *SystemIcons) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return errorx.Invalid("图标 ID 无效")
	}
	result := s.db.WithContext(ctx).Delete(&entity.SystemIcon{}, id)
	if result.Error == nil && result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}
