package model

import (
	"errors"
	"fmt"
	"strings"
)

type Name struct {
	ID     int32    `json:"id" db:"id"`
	Count  int32    `json:"count" db:"count"`
	Text   string   `json:"text" db:"name_text"`
	Type   NameType `json:"type" db:"name_type"`
	Gender Gender   `json:"gender" db:"gender"`
}

func (n Name) Validate() error {
	var errs []error
	if !n.Type.IsValid() {
		errs = append(errs, fmt.Errorf("name_type must be enums %q, got %q", AllGenders, n.Type))
	}
	if !n.Gender.IsValid() {
		errs = append(errs, fmt.Errorf("gender must be enums %q, got %q", AllNameTypes, n.Gender))
	}
	if err := ValidateName(n.Text); err != nil {
		errs = append(errs, fmt.Errorf("name_text: %w", err))
	}
	return errors.Join(errs...)
}

// NormalizeName. Всегда копирует строку.
func NormalizeName(s string) (string, error) {
	// TODO
	return strings.Clone(strings.TrimSpace(s)), nil
}

func ValidateName(s string) error {
	// TODO
	return nil
}
