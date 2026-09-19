package repository_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	checkoutrepository "github.com/Halturshik/TicketAgregator-API/internal/checkout/repository"
	checkoutservice "github.com/Halturshik/TicketAgregator-API/internal/checkout/service"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/orders"
	passengerservice "github.com/Halturshik/TicketAgregator-API/internal/passengers/service"
	"github.com/Halturshik/TicketAgregator-API/internal/payments"
)

func TestPayCommitsStatusBonusesAndPassengerAutosaveAtomically(t *testing.T) {
	passengerJSON, _ := json.Marshal(orders.PassengerSnapshot{
		FirstName: "Иван", LastName: "Иванов", BirthDate: "1990-01-01", IsRussian: true,
	})
	documentJSON, _ := json.Marshal(orders.DocumentSnapshot{
		Type: "internal_passport", Number: "1234567890", VerificationStatus: "verified",
		LastCheckedAt: "2026-08-22T12:00:00Z",
	})
	future := time.Now().UTC().Add(time.Hour)
	conn := &scriptedConn{steps: []dbStep{
		queryStep("FROM orders", []string{"id", "user_id", "guest_payment_token", "status", "payable_amount", "bonus_spent", "bonus_earned", "expires_at"}, []driver.Value{int64(42), int64(7), nil, "created", int64(5000), int64(100), int64(102), future}),
		queryStep("INSERT INTO payments", []string{"id"}, []driver.Value{int64(11)}),
		execStep("UPDATE orders SET status = $2", 1),
		execStep("UPDATE tickets SET status = $2", 2),
		execStep("pg_advisory_xact_lock", 1),
		queryStep("SET bonus_points = bonus_points -", []string{"bonus_points"}, []driver.Value{int64(150)}),
		execStep("INSERT INTO bonus_transactions", 1),
		queryStep("SELECT bonus_points, bonus_debt", []string{"bonus_points", "bonus_debt"}, []driver.Value{int64(150), int64(0)}),
		execStep("SET bonus_points = $1, bonus_debt = $2", 1),
		execStep("INSERT INTO bonus_transactions", 1),
		queryStep("FROM order_passengers", []string{
			"passenger_source", "saved_passenger_id", "saved_document_id", "save_changes",
			"passenger_snapshot", "document_snapshot", "document_fingerprint",
		}, []driver.Value{"new", nil, nil, false, passengerJSON, documentJSON, "fingerprint"}),
		emptyQueryStep("JOIN documents d"),
		queryStep("INSERT INTO saved_passengers", []string{"id"}, []driver.Value{int64(91)}),
		execStep("INSERT INTO documents", 1),
		queryStep("SELECT bonus_points FROM users", []string{"bonus_points"}, []driver.Value{int64(252)}),
	}}
	db := openScriptedDB(t, conn)
	repo := checkoutrepository.NewRepository(db)
	service := checkoutservice.NewService(repo, successfulPaymentProvider(), passengerservice.NewService(nil))
	userID := 7

	out, err := service.Pay(context.Background(), &userID, 42, "")
	if err != nil {
		t.Fatalf("Pay() error = %v", err)
	}
	if out.PaymentID != 11 || out.BonusBalance == nil || *out.BonusBalance != 252 {
		t.Fatalf("Pay() result = payment:%d balance:%v", out.PaymentID, out.BonusBalance)
	}
	conn.assertComplete(t, true)
}

func TestPayRollsBackEverythingWhenPassengerAutosaveFails(t *testing.T) {
	future := time.Now().UTC().Add(time.Hour)
	conn := &scriptedConn{steps: []dbStep{
		queryStep("FROM orders", []string{"id", "user_id", "guest_payment_token", "status", "payable_amount", "bonus_spent", "bonus_earned", "expires_at"}, []driver.Value{int64(43), int64(7), nil, "created", int64(5000), int64(0), int64(0), future}),
		queryStep("INSERT INTO payments", []string{"id"}, []driver.Value{int64(15)}),
		execStep("UPDATE orders SET status = $2", 1),
		execStep("UPDATE tickets SET status = $2", 2),
		execStep("pg_advisory_xact_lock", 1),
		queryStep("FROM order_passengers", []string{
			"passenger_source", "saved_passenger_id", "saved_document_id", "save_changes",
			"passenger_snapshot", "document_snapshot", "document_fingerprint",
		}, []driver.Value{"new", nil, nil, false, []byte(`{"broken"`), []byte(`{}`), "fingerprint"}),
	}}
	db := openScriptedDB(t, conn)
	repo := checkoutrepository.NewRepository(db)
	service := checkoutservice.NewService(repo, successfulPaymentProvider(), passengerservice.NewService(nil))
	userID := 7

	_, err := service.Pay(context.Background(), &userID, 43, "")
	if err == nil || !strings.Contains(err.Error(), "unmarshal passenger snapshot") {
		t.Fatalf("Pay() error = %v, want autosave failure", err)
	}
	conn.assertComplete(t, false)
	if !conn.rolledBack {
		t.Fatal("transaction was not rolled back")
	}
}

func TestPayFailureDoesNotChangeOrderOrAutosavePassengers(t *testing.T) {
	future := time.Now().UTC().Add(time.Hour)
	conn := &scriptedConn{steps: []dbStep{
		queryStep("FROM orders", []string{"id", "user_id", "guest_payment_token", "status", "payable_amount", "bonus_spent", "bonus_earned", "expires_at"}, []driver.Value{int64(8), nil, "guest-token", "created", int64(2000), int64(0), int64(0), future}),
		queryStep("INSERT INTO payments", []string{"id"}, []driver.Value{int64(12)}),
	}}
	db := openScriptedDB(t, conn)
	repo := checkoutrepository.NewRepository(db)
	service := checkoutservice.NewService(repo, failedPaymentProvider(), passengerservice.NewService(nil))

	out, err := service.Pay(context.Background(), nil, 8, "guest-token")
	if err != nil {
		t.Fatalf("Pay() error = %v", err)
	}
	if out.PaymentID != 12 || out.Status != payments.StatusFailed {
		t.Fatalf("payment output = %+v, want failed payment 12", out)
	}
	conn.assertComplete(t, true)
}

func TestPayExpiresOrderUnderSameTransactionLock(t *testing.T) {
	past := time.Now().UTC().Add(-time.Minute)
	conn := &scriptedConn{steps: []dbStep{
		queryStep("FROM orders", []string{"id", "user_id", "guest_payment_token", "status", "payable_amount", "bonus_spent", "bonus_earned", "expires_at"}, []driver.Value{int64(9), nil, "guest-token", "created", int64(2000), int64(0), int64(0), past}),
		execStep("SET status = $2", 1),
		execStep("SET status = $2", 2),
	}}
	db := openScriptedDB(t, conn)
	repo := checkoutrepository.NewRepository(db)
	service := checkoutservice.NewService(repo, successfulPaymentProvider(), passengerservice.NewService(nil))

	_, err := service.Pay(context.Background(), nil, 9, "guest-token")
	if !errors.Is(err, apierror.ErrOrderExpired) {
		t.Fatalf("Pay() error = %v, want expired", err)
	}
	conn.assertComplete(t, true)
}

func TestPayRollsBackWhenOrderStatusWasNotUpdated(t *testing.T) {
	future := time.Now().UTC().Add(time.Hour)
	conn := &scriptedConn{steps: []dbStep{
		queryStep("FROM orders", []string{"id", "user_id", "guest_payment_token", "status", "payable_amount", "bonus_spent", "bonus_earned", "expires_at"}, []driver.Value{int64(10), nil, "guest-token", "created", int64(2000), int64(0), int64(0), future}),
		queryStep("INSERT INTO payments", []string{"id"}, []driver.Value{int64(20)}),
		execStep("UPDATE orders SET status = $2", 0),
	}}
	db := openScriptedDB(t, conn)
	repo := checkoutrepository.NewRepository(db)
	service := checkoutservice.NewService(repo, successfulPaymentProvider(), passengerservice.NewService(nil))

	_, err := service.Pay(context.Background(), nil, 10, "guest-token")
	if err == nil || !strings.Contains(err.Error(), "affected rows") {
		t.Fatalf("Pay() error = %v, want affected rows failure", err)
	}
	conn.assertComplete(t, false)
	if !conn.rolledBack {
		t.Fatal("transaction was not rolled back")
	}
}

func successfulPaymentProvider() payments.Provider {
	return payments.ProviderFunc(func(context.Context) (bool, error) { return true, nil })
}

func failedPaymentProvider() payments.Provider {
	return payments.ProviderFunc(func(context.Context) (bool, error) { return false, nil })
}

type dbStep struct {
	kind     string
	contains string
	columns  []string
	rows     [][]driver.Value
	affected int64
}

func queryStep(contains string, columns []string, row []driver.Value) dbStep {
	return dbStep{kind: "query", contains: contains, columns: columns, rows: [][]driver.Value{row}}
}

func emptyQueryStep(contains string) dbStep {
	return dbStep{kind: "query", contains: contains, columns: []string{"id"}}
}

func execStep(contains string, affected int64) dbStep {
	return dbStep{kind: "exec", contains: contains, affected: affected}
}

var scriptedDriverID atomic.Int64

func openScriptedDB(t *testing.T, conn *scriptedConn) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("scripted-postgres-%d", scriptedDriverID.Add(1))
	sql.Register(name, &scriptedDriver{conn: conn})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("open scripted db: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

type scriptedDriver struct{ conn *scriptedConn }

func (d *scriptedDriver) Open(string) (driver.Conn, error) { return d.conn, nil }

type scriptedConn struct {
	steps      []dbStep
	position   int
	committed  bool
	rolledBack bool
	closed     bool
}

func (c *scriptedConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not supported")
}
func (c *scriptedConn) Close() error { c.closed = true; return nil }
func (c *scriptedConn) Begin() (driver.Tx, error) {
	return &scriptedTx{conn: c}, nil
}
func (c *scriptedConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return &scriptedTx{conn: c}, nil
}
func (c *scriptedConn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	step, err := c.next("query", query)
	if err != nil {
		return nil, err
	}
	return &scriptedRows{columns: step.columns, rows: step.rows}, nil
}
func (c *scriptedConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	step, err := c.next("exec", query)
	if err != nil {
		return nil, err
	}
	return driver.RowsAffected(step.affected), nil
}
func (c *scriptedConn) next(kind string, query string) (dbStep, error) {
	if c.position >= len(c.steps) {
		return dbStep{}, fmt.Errorf("unexpected %s: %s", kind, compactSQL(query))
	}
	step := c.steps[c.position]
	if step.kind != kind || !strings.Contains(query, step.contains) {
		return dbStep{}, fmt.Errorf(
			"step %d: got %s %q, want %s containing %q",
			c.position, kind, compactSQL(query), step.kind, step.contains,
		)
	}
	c.position++
	return step, nil
}
func (c *scriptedConn) assertComplete(t *testing.T, committed bool) {
	t.Helper()
	if c.position != len(c.steps) {
		t.Fatalf("executed %d of %d database steps", c.position, len(c.steps))
	}
	if c.committed != committed {
		t.Fatalf("committed = %t, want %t", c.committed, committed)
	}
}

type scriptedTx struct{ conn *scriptedConn }

func (tx *scriptedTx) Commit() error   { tx.conn.committed = true; return nil }
func (tx *scriptedTx) Rollback() error { tx.conn.rolledBack = true; return nil }

type scriptedRows struct {
	columns []string
	rows    [][]driver.Value
	index   int
}

func (r *scriptedRows) Columns() []string { return r.columns }
func (r *scriptedRows) Close() error      { return nil }
func (r *scriptedRows) Next(dest []driver.Value) error {
	if r.index >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.index])
	r.index++
	return nil
}

func compactSQL(query string) string { return strings.Join(strings.Fields(query), " ") }
