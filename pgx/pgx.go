package pgx

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/knadh/goyesql/v2"
)

// PreparedQuery wraps a pgx query with the pool for easy execution.
// It provides methods similar to pgx.Conn but uses the connection pool.
type PreparedQuery struct {
	pool  *pgxpool.Pool
	query string
	ctx   context.Context
}

// Query executes a query that returns rows.
func (pq *PreparedQuery) Query(args ...interface{}) (pgx.Rows, error) {
	return pq.pool.Query(pq.ctx, pq.query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
func (pq *PreparedQuery) QueryRow(args ...interface{}) pgx.Row {
	return pq.pool.QueryRow(pq.ctx, pq.query, args...)
}

// Exec executes a query without returning any rows.
func (pq *PreparedQuery) Exec(args ...interface{}) (pgconn.CommandTag, error) {
	return pq.pool.Exec(pq.ctx, pq.query, args...)
}

// ScanToStruct prepares a given set of Queries and assigns the resulting
// query strings to the fields of a given struct, matching based on the name
// in the `query` tag in the struct field names.
func ScanToStruct(obj interface{}, q goyesql.Queries, pool *pgxpool.Pool) error {
	return ScanWithContext(context.Background(), obj, q, pool)
}

// ScanWithContext prepares a given set of Queries and assigns the resulting
// query strings to the fields of a given struct with context support.
// It matches fields based on the name in the `query` tag in the struct field names.
func ScanWithContext(ctx context.Context, obj interface{}, q goyesql.Queries, pool *pgxpool.Pool) error {
	ob := reflect.ValueOf(obj)
	if ob.Kind() == reflect.Ptr {
		ob = ob.Elem()
	}

	if ob.Kind() != reflect.Struct {
		return fmt.Errorf("Failed to apply SQL statements to struct. Non struct type: %T", ob)
	}

	// Go through every field in the struct and look for it in the Args map.
	for i := 0; i < ob.NumField(); i++ {
		f := ob.Field(i)

		if f.IsValid() {
			if tag := ob.Type().Field(i).Tag.Get("query"); tag != "" && tag != "-" {
				// Extract the value of the `query` tag.
				var (
					tg   = strings.Split(tag, ",")
					name string
				)
				if len(tg) == 2 {
					if tg[0] != "-" && tg[0] != "" {
						name = tg[0]
					}
				} else {
					name = tg[0]
				}

				// Query name found in the field tag is not in the map.
				if _, ok := q[name]; !ok {
					return fmt.Errorf("query '%s' not found in query map", name)
				}

				if !f.CanSet() {
					return fmt.Errorf("query field '%s' is unexported", ob.Type().Field(i).Name)
				}

				switch f.Type() {
				case reflect.TypeOf(""):
					// Unprepared SQL query string.
					f.Set(reflect.ValueOf(q[name].Query))
				case reflect.TypeOf((*PreparedQuery)(nil)):
					// Wrapped pgx query with pool.
					pq := &PreparedQuery{
						pool:  pool,
						query: q[name].Query,
						ctx:   ctx,
					}
					f.Set(reflect.ValueOf(pq))
				}
			}
		}
	}

	return nil
}
