package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) Run() error {
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to init migrations table: %v", err)
	}

	// Получаем список примененных миграций
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	// Применяем миграции по очереди
	for _, migration := range migrations {
		if _, ok := applied[migration.Name]; !ok {
			if err := m.applyMigration(migration); err != nil {
				return fmt.Errorf("migration %s failed: %v", migration.Name, err)
			}
		}
	}

	return nil
}

func (m *Migrator) applyMigration(migration Migration) error {
	log.Printf("Applying migration: %s", migration.Name)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Выполняем каждую команду миграции
	for _, cmd := range migration.Commands {
		if _, err := tx.ExecContext(ctx, cmd); err != nil {
			return fmt.Errorf("failed to execute command:\n%s\nError: %v", cmd, err)
		}
	}

	// Фиксируем миграцию
	if _, err := tx.ExecContext(
		ctx,
		"INSERT INTO migrations (name) VALUES (?)",
		migration.Name,
	); err != nil {
		return fmt.Errorf("failed to record migration: %v", err)
	}

	return tx.Commit()
}

func (m *Migrator) createMigrationsTable() error {
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	return err
}

func (m *Migrator) getAppliedMigrations() (map[string]struct{}, error) {
	rows, err := m.db.Query("SELECT name FROM migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %v", err)
	}
	defer rows.Close()

	applied := make(map[string]struct{})
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = struct{}{}
	}

	return applied, nil
}
