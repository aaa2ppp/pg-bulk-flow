package main

import (
	"cmp"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"pg-bulk-flow/internal/config"
	"pg-bulk-flow/internal/database"
	"pg-bulk-flow/internal/inserter"
	"pg-bulk-flow/internal/inserter/copyfrom"
	"pg-bulk-flow/internal/inserter/pgxbatch"
	"pg-bulk-flow/internal/inserter/unnestbatch"
	"pg-bulk-flow/internal/logger"
	"pg-bulk-flow/internal/model"
	"pg-bulk-flow/internal/parser"
	"pg-bulk-flow/internal/profiling"
	"pg-bulk-flow/internal/scanner"
	"pg-bulk-flow/internal/strutils"

	"github.com/joho/godotenv"
)

const (
	defaultBatchSize = 1000
	defaulTimeout    = 1 * time.Minute // чтобы не ждать вечность
)

var (
	inputFile = flag.String("i", "", "Input file ($INPUT_FILE, default: stdin)")
	nameType  = flag.String("type", "", "Type of names to insert ($NAME_TYPE). Available values: "+strutils.Join(model.AllNameTypes, ", "))
	timeout   = flag.Duration("timeout", defaulTimeout, "Maximum processing duration (0 or negative means no timeout)")
	method    = flag.String("method", "copyfrom", "Insert method to use: copyfrom, pgxbatch or unnestbatch")
	batchSize = flag.Int("batch", defaultBatchSize, "Number of records per batch insert")
	truncate  = flag.Bool("truncate", false, "Clear the table before inserting new records")
	pipeline  = flag.Bool("pipeline", false, "Enable concurrent scanning and inserting for better performance")
)

func main() {
	godotenv.Load()
	flag.Parse()
	cfg := loadConfig()
	logger.SetupDefault(cfg.Log)
	os.Exit(run(cfg))
}

func loadConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("can't load config: %v", err)
	}

	if *batchSize <= 0 {
		fmt.Fprintln(os.Stderr, "batch size must be positive")
		flag.PrintDefaults()
		os.Exit(1)
	}

	supportedMethods := map[string]bool{
		"copyfrom":    true,
		"pgxbatch":    true,
		"unnestbatch": true,
	}
	if !supportedMethods[*method] {
		fmt.Fprintf(os.Stderr, "invalid method: %s\n", *method)
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *inputFile != "" {
		cfg.InputFile = *inputFile
	}

	if *nameType != "" {
		if v, err := model.ParseNameType(*nameType); err != nil {
			fmt.Fprintf(os.Stderr, "invalid name type: %v\n", err)
			flag.PrintDefaults()
			os.Exit(1)
		} else {
			cfg.NameType = v
		}
	}

	return cfg
}

type totalStats struct {
	Elapsed  time.Duration `json:"elapsed,omitempty"`
	Parser   parser.Stats  `json:"parser,omitempty"`
	Scanner  scanner.Stats `json:"scanner,omitempty"`
	Inserted int64         `json:"inserted,omitempty"`
}

type insertConfig struct {
	Input     string         `json:"input,omitempty"`
	NameType  model.NameType `json:"name_type,omitempty"`
	Method    string         `json:"method,omitempty"`
	Pipeline  bool           `json:"pipeline,omitempty"`
	BatchSize int            `json:"batch_size,omitempty"`
	Timeout   time.Duration  `json:"timeout,omitempty"`
}

func run(cfg *config.Config) int {
	input := os.Stdin
	if cfg.InputFile != "" {
		var err error
		input, err = os.Open(cfg.InputFile)
		if err != nil {
			slog.Error("open file failed", "error", err)
			return 1
		}
		defer input.Close()
	}

	conn, err := database.Connect(cfg.DB)
	if err != nil {
		slog.Error("database connect failed", "error", err)
		return 1
	}
	defer conn.Close(context.Background())

	if *truncate {
		if _, err := conn.Exec(context.Background(), `TRUNCATE TABLE names`); err != nil {
			slog.Error("truncate table failed", "error", err)
			return 1
		}
	}

	parser := new(parser.Parser)
	scanner := scanner.New(input, cfg.NameType, parser)

	var inserter inserter.Inserter
	switch *method {
	case "copyfrom":
		inserter = copyfrom.New(conn)
	case "pgxbatch":
		inserter = pgxbatch.New(conn, *batchSize)
	case "unnestbatch":
		inserter = unnestbatch.New(conn, *batchSize)
	default:
		slog.Error("unknown insert method", "method", *method)
		return 1
	}

	insert := inserter.Insert
	if *pipeline {
		insert = inserter.InsertWithPipeline
	}

	ctx := context.Background()
	if *timeout >= 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}

	var (
		elapsed time.Duration
		count   int64
		insErr  error
	)

	profiling.Do(func() {
		start := time.Now()
		count, insErr = insert(ctx, scanner.Scan(ctx))
		elapsed = time.Since(start)
	})

	if err := scanner.Err(); err != nil {
		slog.Error("scan failed", "error", err)
		return 1
	}

	if insErr != nil {
		slog.Error("insert failed", "error", insErr)
		return 1
	}

	results := struct {
		Config insertConfig `json:"config,omitempty"`
		Stats  totalStats   `json:"stats,omitempty"`
	}{
		Config: insertConfig{
			Input:     cmp.Or(*inputFile, "stdin"),
			NameType:  cfg.NameType,
			Method:    *method,
			BatchSize: *batchSize,
			Pipeline:  *pipeline,
			Timeout:   *timeout / time.Millisecond, // to milliseconds
		},
		Stats: totalStats{
			Elapsed:  elapsed / time.Millisecond, // to milliseconds
			Parser:   parser.Stats(),
			Scanner:  scanner.Stats(),
			Inserted: count,
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(results); err != nil {
		slog.Error("encode stats failded", "error", err)
		return 1
	}

	return 0
}
