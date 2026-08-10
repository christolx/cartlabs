package database

import (
	"context"
	"strings"
	"testing"
)

func TestEnsureDatabaseRejectsSystemTargetsBeforeConnecting(t *testing.T) {
	for _, databaseName := range []string{"postgres", "template0", "template1"} {
		err := EnsureDatabase(context.Background(), "postgres://cartlabs:cartlabs@127.0.0.1:1/"+databaseName+"?sslmode=disable")
		if err == nil || !strings.Contains(err.Error(), "refuse unsafe database") {
			t.Fatalf("database=%s error=%v", databaseName, err)
		}
	}
}
