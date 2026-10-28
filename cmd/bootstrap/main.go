package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"web-api/internal/config"
	"web-api/internal/database"
	"web-api/internal/models"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("configuration error: ", err)
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		log.Fatal("database connection failed: ", err)
	}
	defer db.Close()

	// Блокировка предотвращает параллельную
	// первоначальную инициализацию.
	tx, err := db.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		LOCK TABLE app_users IN EXCLUSIVE MODE
	`)
	if err != nil {
		log.Fatal(err)
	}

	var count int64
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM app_users
	`).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count > 0 {
		log.Fatal("bootstrap unavailable: users already exist")
	}

	fmt.Print("Admin username: ")

	var username string
	if _, err := fmt.Scanln(&username); err != nil {
		log.Fatal(err)
	}

	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 50 {
		log.Fatal("username must contain 3-50 characters")
	}

	admin := models.User{
		Username: username,
		Role:     models.RoleAdmin,
	}

	if err := admin.Validate(); err != nil {
		log.Fatal("invalid admin user: ", err)
	}

	fmt.Print("Admin password: ")

	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal("cannot read password: ", err)
	}
	fmt.Println()

	if len(password) < 12 || len(password) > 72 {
		log.Fatal("password must contain 12-72 bytes")
	}

	hash, err := bcrypt.GenerateFromPassword(
		password,
		bcrypt.DefaultCost,
	)
	if err != nil {
		log.Fatal(err)
	}

	var id int64
	err = tx.QueryRow(ctx, `
		INSERT INTO app_users (
			username,
			password_hash,
			role
		)
		VALUES ($1, $2, $3)
		RETURNING id
	`, admin.Username, string(hash), admin.Role).Scan(&id)

	if errors.Is(err, pgx.ErrNoRows) {
		log.Fatal("failed to create admin")
	}
	if err != nil {
		log.Fatal(err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Printf(
		"Administrator created: id=%d, username=%s\n",
		id,
		username,
	)
}
