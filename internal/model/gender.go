package model

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

// Gender represents person's gender identity
//
//go:generate stringer -type Gender -linecomment
type Gender int8

const (
	GenderUnknown Gender = iota // unknown
	GenderMale                  // male
	GenderFemale                // female
)

var AllGenders = []Gender{GenderUnknown, GenderMale, GenderFemale}

func (g Gender) IsValid() bool {
	return slices.Contains(AllGenders, g)
}

// TextValue implements pgtype.TextValuer.
func (g Gender) TextValue() (pgtype.Text, error) {
	if !g.IsValid() {
		return pgtype.Text{}, fmt.Errorf("invalid gender %v", g)
	}
	return pgtype.Text{String: g.String(), Valid: true}, nil
}

// ScanText implements pgtype.TextScanner.
func (g *Gender) ScanText(v pgtype.Text) error {
	if !v.Valid {
		*g = GenderUnknown // null -> unknown
		return nil
	}
	gv, err := ParseGender(v.String)
	if err != nil {
		return err
	}
	*g = gv
	return nil
}

// MarshalJSON implements json.Marshaler.
func (g Gender) MarshalJSON() ([]byte, error) {
	return []byte(g.String()), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *Gender) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	v, err := ParseGender(s)
	if err != nil {
		return err
	}
	*g = v
	return nil
}

func ParseGender(s string) (Gender, error) {
	s = strings.TrimSpace(s)
	switch {
	case s == "m", s == "M", strings.EqualFold(s, "male"):
		return GenderMale, nil
	case s == "f", s == "F", strings.EqualFold(s, "female"):
		return GenderFemale, nil
	case s == "", s == "u", s == "U", strings.EqualFold(s, "unknown"):
		return GenderUnknown, nil
	}
	return GenderUnknown, fmt.Errorf("invalid gender %q", s)
}

func DetectGender(name string, nameType NameType) Gender {
	name = strings.ToLower(strings.TrimSpace(name))
	switch nameType {
	case NameTypeSurname:
		if strings.HasSuffix(name, "ов") ||
			strings.HasSuffix(name, "eв") ||
			strings.HasSuffix(name, "ёв") ||
			strings.HasSuffix(name, "ин") ||
			strings.HasSuffix(name, "ын") ||
			strings.HasSuffix(name, "ой") ||
			strings.HasSuffix(name, "ий") ||
			strings.HasSuffix(name, "ый") ||
			false {
			return GenderMale
		}
		if strings.HasSuffix(name, "ова") ||
			strings.HasSuffix(name, "eва") ||
			strings.HasSuffix(name, "ёва") ||
			strings.HasSuffix(name, "вна") ||
			strings.HasSuffix(name, "ина") ||
			strings.HasSuffix(name, "ына") ||
			strings.HasSuffix(name, "ая") ||
			strings.HasSuffix(name, "яя") ||
			false {
			return GenderFemale
		}

	case NameTypePatronymic:
		if strings.HasSuffix(name, "ович") ||
			strings.HasSuffix(name, "евич") ||
			strings.HasSuffix(name, "ич") {
			return GenderMale
		}
		if strings.HasSuffix(name, "овна") ||
			strings.HasSuffix(name, "евна") ||
			strings.HasSuffix(name, "ична") ||
			strings.HasSuffix(name, "инична") {
			return GenderMale
		}
	}
	return GenderUnknown
}

var (
	_ json.Unmarshaler   = (*Gender)(nil)
	_ json.Marshaler     = Gender(1)
	_ pgtype.TextScanner = new(Gender)
	_ pgtype.TextValuer  = Gender(0)
)
