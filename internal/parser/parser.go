package parser

import (
	"context"
	"errors"
	"fmt"
	"math"

	"pg-bulk-flow/internal/model"
	"pg-bulk-flow/internal/scanner"
)

//go:generate easyjson .

// inputRecord cтруктура входящих данных. Для производительности используется eastjson
// с тегом nocopy. Использование nocopy здесь безопасно поскольку:
//   - Count в конечном счете копируется в int32 (numberLong — кастомный тип для парсинга
//     сложных объектов в int64, реализует json.Unmarshaler);
//   - Text неявно клонируется функцией model.NormalizeName;
//   - Gender парсится в model.Gender (byte);
//
// Остальные поля только проверяются на zero-value.
//
//easyjson:json
type inputRecord struct {
	Count  numberLong `json:"count,omitempty,nocopy"`
	Text   string     `json:"text,omitempty,nocopy"`
	Gender string     `json:"gender,omitempty,nocopy"`
	FName  string     `json:"fname,omitempty,nocopy"`
	FForm  string     `json:"f_form,omitempty,nocopy"`
	MForm  string     `json:"m_form,omitempty,nocopy"`
	Ethnic []string   `json:"ethnic,omitempty,nocopy"`
}

type Stats struct {
	InvalidJSON   int `json:"invalid_json,omitempty"`
	EmptyFields   int `json:"empty_fields,omitempty"`
	InvalidName   int `json:"invalid_name,omitempty"`
	InvalidGender int `json:"invalid_gender,omitempty"`
	InvalidCount  int `json:"invalid_count,omitempty"`
}

// Parse парсит входные данные в model.Name.
// Парсер НЕ потокобезопасен. Создавайте новый для каждой горутины.
type Parser struct {
	stats Stats
}

func (p *Parser) Stats() Stats {
	return p.stats
}

func (p *Parser) Parse(ctx context.Context, data []byte) (model.Name, error) {
	var rec inputRecord
	if err := rec.UnmarshalJSON(data); err != nil {
		p.stats.InvalidJSON++
		return model.Name{}, err
	}

	if rec.Count == 1 && rec.Gender == "" && rec.FName == "" && rec.FForm == "" && rec.MForm == "" && rec.Ethnic == nil {
		// Запись игнорируется: все поля пусты или nil (скорее всего мусор).
		// Полезные данные обычно содержат хотя бы одно дополнительное поле.
		p.stats.EmptyFields++
		return model.Name{}, errors.New("too little data")
	}

	// Остальное — ответственность валидатора
	name, err := model.NormalizeName(rec.Text)
	if err != nil {
		p.stats.InvalidName++
		return model.Name{}, fmt.Errorf("invalid text: %w", err)
	}

	gender, err := model.ParseGender(rec.Gender)
	if err != nil {
		p.stats.InvalidGender++
		return model.Name{}, fmt.Errorf("invalid gender: %w", err)
	}

	if !(0 < rec.Count && rec.Count <= math.MaxInt32) {
		p.stats.InvalidCount++
		return model.Name{}, fmt.Errorf("count must be [1..%d]", math.MaxInt32)
	}

	return model.Name{
		Text:   name,
		Gender: gender,
		Count:  int32(rec.Count),
	}, nil
}

var _ scanner.Parser = &Parser{}
