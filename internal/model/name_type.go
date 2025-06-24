package model

import (
	"fmt"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

// NameType represents person's name type (firstname/surname/patronymic)
//
//go:generate stringer -type NameType -linecomment -output name_type_string.go
type NameType int8

const (
	_                  NameType = iota
	NameTypeName                // name
	NameTypeSurname             // surname
	NameTypePatronymic          // patronymic
)

var AllNameTypes = []NameType{NameTypeName, NameTypeSurname, NameTypePatronymic}

func (t NameType) IsValid() bool {
	return slices.Contains(AllNameTypes, t)
}

// ScanText implements pgtype.TextScanner.
func (t *NameType) ScanText(v pgtype.Text) error {
	tv, err := ParseNameType(v.String)
	if err != nil {
		return err
	}
	*t = tv
	return nil
}

// TextValue implements pgtype.TextValuer.
func (t NameType) TextValue() (pgtype.Text, error) {
	if !t.IsValid() {
		return pgtype.Text{}, fmt.Errorf("invalid name type value %v", t)
	}
	return pgtype.Text{String: t.String(), Valid: true}, nil
}

func ParseNameType(s string) (NameType, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "name", "firstname":
		return NameTypeName, nil
	case "surname", "lastname":
		return NameTypeSurname, nil
	case "patronymic", "midname", "middlename":
		return NameTypePatronymic, nil
	}
	return 0, fmt.Errorf("unknown name type %q", s)
}

var (
	_ pgtype.TextScanner = new(NameType)
	_ pgtype.TextValuer  = NameType(0)
)
