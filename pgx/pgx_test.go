package pgx

import (
	"context"
	"testing"

	"github.com/knadh/goyesql/v2"
)

func TestScanToStruct(t *testing.T) {
	// Test basic query map
	queries := goyesql.Queries{
		"get_user": &goyesql.Query{
			Query: "SELECT * FROM users WHERE id = $1",
			Tags:  map[string]string{},
		},
		"list_users": &goyesql.Query{
			Query: "SELECT * FROM users",
			Tags:  map[string]string{},
		},
	}

	type TestQueries struct {
		GetUser   string `query:"get_user"`
		ListUsers string `query:"list_users"`
	}

	var q TestQueries
	// Note: ScanToStruct with nil pool only works for string fields
	err := ScanToStruct(&q, queries, nil)
	if err != nil {
		t.Fatalf("ScanToStruct failed: %v", err)
	}

	if q.GetUser != queries["get_user"].Query {
		t.Errorf("Expected GetUser = %s, got %s", queries["get_user"].Query, q.GetUser)
	}

	if q.ListUsers != queries["list_users"].Query {
		t.Errorf("Expected ListUsers = %s, got %s", queries["list_users"].Query, q.ListUsers)
	}
}

func TestScanWithContext(t *testing.T) {
	queries := goyesql.Queries{
		"get_user": &goyesql.Query{
			Query: "SELECT * FROM users WHERE id = $1",
			Tags:  map[string]string{},
		},
	}

	type TestQueries struct {
		GetUser string `query:"get_user"`
	}

	var q TestQueries
	ctx := context.Background()
	err := ScanWithContext(ctx, &q, queries, nil)
	if err != nil {
		t.Fatalf("ScanWithContext failed: %v", err)
	}

	if q.GetUser != queries["get_user"].Query {
		t.Errorf("Expected GetUser = %s, got %s", queries["get_user"].Query, q.GetUser)
	}
}

func TestScanToStructMissingQuery(t *testing.T) {
	queries := goyesql.Queries{
		"get_user": &goyesql.Query{
			Query: "SELECT * FROM users WHERE id = $1",
			Tags:  map[string]string{},
		},
	}

	type TestQueries struct {
		GetUser     string `query:"get_user"`
		MissingUser string `query:"missing_user"`
	}

	var q TestQueries
	err := ScanToStruct(&q, queries, nil)
	if err == nil {
		t.Fatal("Expected error for missing query, got nil")
	}
}

func TestScanToStructUnexportedField(t *testing.T) {
	queries := goyesql.Queries{
		"get_user": &goyesql.Query{
			Query: "SELECT * FROM users WHERE id = $1",
			Tags:  map[string]string{},
		},
	}

	type TestQueries struct {
		getUser string `query:"get_user"`
	}

	var q TestQueries
	err := ScanToStruct(&q, queries, nil)
	if err == nil {
		t.Fatal("Expected error for unexported field, got nil")
	}
}

func TestScanToStructTagParsing(t *testing.T) {
	queries := goyesql.Queries{
		"get_user": &goyesql.Query{
			Query: "SELECT * FROM users WHERE id = $1",
			Tags:  map[string]string{},
		},
		"list_users": &goyesql.Query{
			Query: "SELECT * FROM users",
			Tags:  map[string]string{},
		},
	}

	type TestQueries struct {
		GetUser   string `query:"get_user,omitempty"`
		ListUsers string `query:"list_users"`
		Ignored   string `query:"-"`
	}

	var q TestQueries
	err := ScanToStruct(&q, queries, nil)
	if err != nil {
		t.Fatalf("ScanToStruct failed: %v", err)
	}

	if q.GetUser != queries["get_user"].Query {
		t.Errorf("Expected GetUser = %s, got %s", queries["get_user"].Query, q.GetUser)
	}

	if q.ListUsers != queries["list_users"].Query {
		t.Errorf("Expected ListUsers = %s, got %s", queries["list_users"].Query, q.ListUsers)
	}

	if q.Ignored != "" {
		t.Errorf("Expected Ignored to be empty, got %s", q.Ignored)
	}
}

func TestScanToStructNonStruct(t *testing.T) {
	queries := goyesql.Queries{
		"get_user": &goyesql.Query{
			Query: "SELECT * FROM users WHERE id = $1",
			Tags:  map[string]string{},
		},
	}

	var q string
	err := ScanToStruct(&q, queries, nil)
	if err == nil {
		t.Fatal("Expected error for non-struct type, got nil")
	}
}
