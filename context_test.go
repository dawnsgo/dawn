package dawn

import (
	"testing"
)

func TestContextorInterfaceCompliance(t *testing.T) {
	// Compile-time check is in context.go: var _ Contextor = (*Context)(nil)
	// This test verifies runtime behavior

	ctx := &Context{}

	// All getters should return nil without panic
	if ctx.Logger() != nil {
		t.Fatal("expected nil Logger()")
	}
	if ctx.Configurator() != nil {
		t.Fatal("expected nil Configurator()")
	}
	if ctx.Eventbus() != nil {
		t.Fatal("expected nil Eventbus()")
	}
	if ctx.TaskPool() != nil {
		t.Fatal("expected nil TaskPool()")
	}
	if ctx.Cache() != nil {
		t.Fatal("expected nil Cache()")
	}
	if ctx.LockMaker() != nil {
		t.Fatal("expected nil LockMaker()")
	}
}

func TestGetContextReturnsContextor(t *testing.T) {
	// GetContext should return non-nil Contextor
	ctx := GetContext()
	if ctx == nil {
		t.Fatal("expected non-nil Contextor from GetContext()")
	}

	// Should be the same as Default()
	defaultCtx := Default()
	if ctx != defaultCtx {
		t.Fatal("expected GetContext() to return same instance as Default()")
	}
}

func TestDefaultSingleton(t *testing.T) {
	// Default() should always return the same instance
	d1 := Default()
	d2 := Default()
	if d1 != d2 {
		t.Fatal("expected Default() to return same instance")
	}
}

func TestContextAsContextor(t *testing.T) {
	// Verify that Context can be assigned to Contextor
	var c Contextor = &Context{}
	if c == nil {
		t.Fatal("assignment should succeed")
	}
}
