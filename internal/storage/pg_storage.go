package storage

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/zhigulin-io/word-stat/internal/entity"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type PGStorage struct {
	db *sql.DB
}

func NewPGStorage(config Config) *PGStorage {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.Name,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("cannot create sql.DB:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("failed ping to database:", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("cannot get postgres driver for migration:", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migration",
		config.Name,
		driver,
	)
	if err != nil {
		log.Fatal("cannot prepare for migration:", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal("cannot migrate database:", err)
	}

	return &PGStorage{
		db: db,
	}
}

func (s PGStorage) Close() error {
	return s.db.Close()
}

func (s PGStorage) GetStatsForPost(postId int) []entity.Stat {
	buf := make([]entity.Stat, 0, 100)

	q := "SELECT post_id, word, count " +
		"FROM stats " +
		"WHERE post_id = $1 " +
		"ORDER BY count DESC"

	rows, err := s.db.Query(q, postId)
	if err != nil {
		log.Println("cannot fetch stats from database:", err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		stat := entity.Stat{}

		err = rows.Scan(&stat.PostID, &stat.Word, &stat.Count)
		if err != nil {
			log.Println("cannot convert row to Stat:", err)
		}

		buf = append(buf, stat)
	}

	if err = rows.Err(); err != nil {
		log.Println("something went wrong while iterating stats:", err)
	}

	return buf
}

func (s PGStorage) UpdateStatsForPost(postId int, stats []entity.Stat) {
	qStats := make([]string, 0, len(stats))
	deleteQ := "DELETE FROM stats WHERE post_id = $1"

	for _, s := range stats {
		qStats = append(
			qStats,
			fmt.Sprintf("(%d, '%s', %d)", postId, s.Word, s.Count),
		)
	}

	insertQ := fmt.Sprintf(
		"INSERT INTO stats (post_id, word, count) VALUES %s",
		strings.Join(qStats, ","),
	)

	tx, err := s.db.Begin()
	if err != nil {
		log.Println("cannot create transaction for batch update:", err)
		return
	}

	_, err = tx.Exec(deleteQ, postId)
	if err != nil {
		log.Printf(
			"cannot delete stats for post with id = %d: %s",
			postId,
			err.Error(),
		)
		err = tx.Rollback()
		if err != nil {
			log.Fatal(err)
		}
	}

	_, err = tx.Exec(insertQ)
	if err != nil {
		log.Printf(
			"cannot insert stats for post with id = %d: %s",
			postId,
			err.Error(),
		)
		err = tx.Rollback()
		if err != nil {
			log.Fatal(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println("cannot commit transaction:", err)
	}
}
