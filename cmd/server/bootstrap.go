package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/eithansmith/master-of-games/game"
)

// bootstrapUser creates the single application user from BOOTSTRAP_USER/
// BOOTSTRAP_PASS on first run, if no user exists yet. Both are optional: if
// unset, this is a no-op and assumes a user was already created. Safe to
// leave the env vars set across restarts - it's a no-op once the user exists.
func bootstrapUser(ctx context.Context, store *game.PostgresStore) error {
	username := strings.TrimSpace(os.Getenv("BOOTSTRAP_USER"))
	password := os.Getenv("BOOTSTRAP_PASS")
	if username == "" || password == "" {
		return nil
	}

	_, exists, err := store.GetUserByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("bootstrapUser: %w", err)
	}
	if exists {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("bootstrapUser: hash: %w", err)
	}
	if err := store.CreateUser(ctx, username, string(hash)); err != nil {
		return fmt.Errorf("bootstrapUser: create: %w", err)
	}

	log.Printf("bootstrapped user %q", username)
	return nil
}
