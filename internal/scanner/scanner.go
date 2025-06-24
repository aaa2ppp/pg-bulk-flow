package scanner

import (
	"bufio"
	"context"
	"errors"
	"io"
	"iter"

	"pg-bulk-flow/internal/logger"
	"pg-bulk-flow/internal/model"
)

type Parser interface {
	Parse(ctx context.Context, data []byte) (model.Name, error)
}

type Stats struct {
	Total    int `json:"total,omitempty"`    // общее количество прочитанных записей
	Unparsed int `json:"unparsed,omitempty"` // записи забракованные парсером
	Invalid  int `json:"invalid,omitempty"`  // записи не прошедшие валидацию
}

type Scanner struct {
	reader   io.Reader
	nameType model.NameType
	parser   Parser
	stats    Stats
	err      error
}

func New(r io.Reader, nameType model.NameType, parser Parser) *Scanner {
	return &Scanner{
		reader:   r,
		nameType: nameType,
		parser:   parser,
	}
}

func (p *Scanner) Stats() Stats {
	return p.stats
}

func (p *Scanner) Err() error {
	return p.err
}

var ErrScanFailed = errors.New("scan failed")

func (s *Scanner) Scan(ctx context.Context) iter.Seq[model.Name] {
	log := logger.FromContext(ctx).With("op", "Scan")
	sc := bufio.NewScanner(s.reader)

	return func(yield func(model.Name) bool) {
		var lineNum = 0
		for sc.Scan() {
			lineNum++
			s.stats.Total++

			name, err := s.parser.Parse(ctx, sc.Bytes())
			if err != nil {
				s.stats.Unparsed++
				log.Debug("skip bad line", "error", err, "lineNum", lineNum)
				continue
			}

			name.Type = s.nameType
			if err := name.Validate(); err != nil {
				s.stats.Invalid++
				log.Debug("invalid record", "error", err, "lineNum", lineNum)
				continue
			}

			if !yield(name) {
				log.Warn("scan break", "lineNum", lineNum)
				break
			}
		}

		if err := sc.Err(); err != nil {
			log.Error("scan failed", "error", err, "lineNum", lineNum+1)
			s.err = ErrScanFailed
		}
	}
}
