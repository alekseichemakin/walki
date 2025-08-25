package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
)

type Storage struct {
	db      *pgxpool.Pool
	queries map[string]string
}

func NewStorage() *Storage {
	url := os.Getenv("DATABASE_URL")
	dbpool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	// Загружаем все SQL-запросы
	queries := make(map[string]string)
	files, err := queriesFS.ReadDir("queries")
	if err != nil {
		log.Fatalf("Unable to read queries directory: %v\n", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		content, err := queriesFS.ReadFile("queries/" + file.Name())
		if err != nil {
			log.Fatalf("Unable to read query file %s: %v\n", file.Name(), err)
		}
		queries[file.Name()] = string(content)
	}

	return &Storage{db: dbpool, queries: queries}
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) GetCities() ([]string, error) {
	query := s.queries["get_cities.sql"]
	rows, err := s.db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get cities: %w", err)
	}
	defer rows.Close()

	var cities []string
	for rows.Next() {
		var city string
		if err := rows.Scan(&city); err != nil {
			return nil, fmt.Errorf("failed to scan city: %w", err)
		}
		cities = append(cities, city)
	}

	return cities, nil
}
