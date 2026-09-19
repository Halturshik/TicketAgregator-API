package requestid

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestResolve(t *testing.T) {
	const valid = "2B0D9D2B-0194-4DC6-804B-EC13AB1C6B40"
	if actual := Resolve(valid); actual != "2b0d9d2b-0194-4dc6-804b-ec13ab1c6b40" {
		t.Fatalf("Resolve(valid) = %q", actual)
	}
	generated := Resolve("not-a-uuid")
	if uuid.Validate(generated) != nil {
		t.Fatalf("Resolve(invalid) = %q", generated)
	}
}

func TestContext(t *testing.T) {
	ctx := WithContext(context.Background(), "request-123")
	if actual, ok := FromContext(ctx); !ok || actual != "request-123" {
		t.Fatalf("FromContext() = %q, %t", actual, ok)
	}
	if _, ok := FromContext(context.Background()); ok {
		t.Fatal("empty context unexpectedly contains request id")
	}
}
