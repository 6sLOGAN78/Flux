package database

import (
	"context"
	"errors"
	"testing"
)

type failingPool struct {
	err    error
	closed int
}

func (p *failingPool) Ping(context.Context) error { return p.err }
func (p *failingPool) Close()                     { p.closed++ }

func TestDatabasePingFailureClosesPool(t *testing.T) {
	cause := errors.New("ping failed")
	pool := &failingPool{err: cause}
	err := pingPool(context.Background(), pool)
	if !errors.Is(err, cause) {
		t.Fatalf("ping cause lost: %v", err)
	}
	if pool.closed != 1 {
		t.Fatalf("failed pool closed %d times", pool.closed)
	}
}

func TestDatabaseSuccessfulPingKeepsPool(t *testing.T) {
	pool := &failingPool{}
	if err := pingPool(context.Background(), pool); err != nil {
		t.Fatal(err)
	}
	if pool.closed != 0 {
		t.Fatal("successful construction closed its pool")
	}
}
