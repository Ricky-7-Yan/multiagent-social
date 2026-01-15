package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yourname/multiagent-social/internal/api"
	"github.com/yourname/multiagent-social/internal/persistence"
)

func main() {
	ctx := context.Background()
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/multiagent?sslmode=disable"
	}
	store, err := persistence.NewPostgresStore(ctx, dsn)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer store.Close()

	if len(os.Args) < 2 {
		fmt.Println("usage: cli <command> [args]")
		fmt.Println("commands: create-agent <name>, gen-token <subject> <days>")
		return
	}
	switch os.Args[1] {
	case "create-agent":
		if len(os.Args) < 3 {
			fmt.Println("usage: cli create-agent <name>")
			return
		}
		name := os.Args[2]
		// Create an actual agent in DB
		id, err := store.CreateAgent(ctx, name, "auto-created persona", map[string]interface{}{"mvp": true})
		if err != nil {
			log.Fatalf("create agent: %v", err)
		}
		fmt.Printf("created agent id: %s\n", id)
	case "gen-token":
		// generate JWT for admin / operations
		if len(os.Args) < 4 {
			fmt.Println("usage: cli gen-token <subject> <days> [role]")
			return
		}
		subject := os.Args[2]
		days := os.Args[3]
		ttlDays := 1
		if n, err := fmt.Sscan(days, &ttlDays); n == 0 || err != nil {
			ttlDays = 1
		}
		role := "admin"
		if len(os.Args) >= 5 {
			role = os.Args[4]
		}
		// generate token using api package
		tok, err := api.GenerateToken(subject, role, time.Duration(ttlDays)*24*time.Hour)
		if err != nil {
			log.Fatalf("failed to generate token: %v", err)
		}
		fmt.Println(tok)
	default:
		fmt.Println("unknown command")
	}
}

