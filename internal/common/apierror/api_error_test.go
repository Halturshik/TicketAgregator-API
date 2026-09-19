package apierror

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestWrapPreservesCauseWithoutExposingIt(t *testing.T) {
	cause := errors.New("database password leaked by driver")
	wrapped := Wrap(fmt.Errorf("get user: %w", cause), ErrInternal)

	if !errors.Is(wrapped, ErrInternal) {
		t.Fatal("wrapped error does not match internal error template")
	}
	if !errors.Is(wrapped, cause) {
		t.Fatal("wrapped error does not preserve its cause")
	}
	if !errors.Is(Cause(wrapped), cause) {
		t.Fatal("Cause does not return the internal error chain")
	}

	encoded, err := json.Marshal(wrapped)
	if err != nil {
		t.Fatalf("marshal wrapped error: %v", err)
	}
	if strings.Contains(string(encoded), cause.Error()) || strings.Contains(string(encoded), "get user") {
		t.Fatalf("internal cause was exposed in JSON: %s", encoded)
	}
	if string(encoded) != `{"code":"internal_error","message":"Что-то пошло не так. Повторите попытку позже"}` {
		t.Fatalf("unexpected public JSON: %s", encoded)
	}
}

func TestWrapDoesNotMutateTemplate(t *testing.T) {
	firstCause := errors.New("first")
	secondCause := errors.New("second")
	first := Wrap(firstCause, ErrInternal)
	second := Wrap(secondCause, ErrInternal)

	if first == ErrInternal || second == ErrInternal || first == second {
		t.Fatal("Wrap must create an independent error instance")
	}
	if Cause(ErrInternal) != ErrInternal {
		t.Fatal("global error template was mutated")
	}
	if !errors.Is(Cause(first), firstCause) || !errors.Is(Cause(second), secondCause) {
		t.Fatal("wrapped instances do not retain independent causes")
	}
}

func TestWrapKeepsKnownAPIError(t *testing.T) {
	if wrapped := Wrap(ErrNotFound, ErrInternal); wrapped != ErrNotFound {
		t.Fatal("known API error must be returned unchanged")
	}

	outer := fmt.Errorf("lookup booking: %w", ErrNotFound)
	wrapped := Wrap(outer, ErrInternal)
	if wrapped == ErrNotFound {
		t.Fatal("an outer error context must be retained")
	}
	if !errors.Is(wrapped, ErrNotFound) || Cause(wrapped) != outer {
		t.Fatal("wrapped API error lost its public type or internal context")
	}
}
