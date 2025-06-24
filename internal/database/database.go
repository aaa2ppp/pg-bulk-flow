package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"pg-bulk-flow/internal/config"
)

func Open(cfg config.DB) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectString())
	if err != nil {
		return nil, fmt.Errorf("parse config failed: %w", err)
	}

	// Регистрируем типы для каждого нового соединения
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return registerEnums(conn, "name_type_enum", "gender_enum")
	}

	return pgxpool.NewWithConfig(context.Background(), poolConfig)
}

func Connect(cfg config.DB) (*pgx.Conn, error) {
	connConfig, err := pgx.ParseConfig(cfg.ConnectString())
	if err != nil {
		return nil, fmt.Errorf("parse config failed: %w", err)
	}

	conn, err := pgx.ConnectConfig(context.Background(), connConfig)
	if err != nil {
		return nil, fmt.Errorf("connect to database failed: %w", err)
	}

	if err := registerEnums(conn, "name_type_enum", "gender_enum"); err != nil {
		conn.Close(context.Background())
		return nil, fmt.Errorf("register enums failed: %w", err)
	}
	
	return conn, nil
}

func registerEnums(conn *pgx.Conn, typnames ...string) error {
	for _, typname := range typnames {
		// TODO: just one request is enough!
		if err := registerEnum(conn, typname); err != nil {
			return err
		}
	}
	return nil
}

func registerEnum(conn *pgx.Conn, typname string) error {
	var (
		baseOID  uint32
		arrayOID uint32
	)

	err := conn.QueryRow(context.Background(),
		`SELECT oid, typarray FROM pg_type where typname=$1`, typname).
		Scan(&baseOID, &arrayOID)

	if err != nil {
		return fmt.Errorf("failed to get OIDs for %v: %w", typname, err)
	}

	typ := &pgtype.Type{
		Name:  typname,
		OID:   baseOID,
		Codec: &pgtype.EnumCodec{},
	}

	// Регистрируем базовый тип
	conn.TypeMap().RegisterType(typ)

	// Регистрируем массив
	conn.TypeMap().RegisterType(&pgtype.Type{
		Name:  "_" + typname,
		OID:   arrayOID,
		Codec: &pgtype.ArrayCodec{ElementType: typ},
	})

	return nil
}
