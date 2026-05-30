package core

import "testing"

func TestNewInputHandler(t *testing.T) {
	h := NewInputHandler()
	if h == nil {
		t.Fatal("NewInputHandler returned nil")
	}
	// Empty on construction
	if _, ok := h.GetKey(); ok {
		t.Error("fresh handler should have no keys")
	}
}

func TestInputHandler_HandleAndGet(t *testing.T) {
	h := NewInputHandler()
	h.HandleKey("left")
	h.HandleKey("right")

	k, ok := h.GetKey()
	if !ok || k != "left" {
		t.Errorf("first GetKey: got (%q, %v), want (left, true)", k, ok)
	}
	k, ok = h.GetKey()
	if !ok || k != "right" {
		t.Errorf("second GetKey: got (%q, %v), want (right, true)", k, ok)
	}
	// Queue drained
	if _, ok := h.GetKey(); ok {
		t.Error("queue should be empty after draining")
	}
}

func TestInputHandler_Clear(t *testing.T) {
	h := NewInputHandler()
	h.HandleKey("up")
	h.HandleKey("down")
	h.Clear()
	if _, ok := h.GetKey(); ok {
		t.Error("Clear should empty the key queue")
	}
}

func TestInputHandler_GetKey_Empty(t *testing.T) {
	h := NewInputHandler()
	k, ok := h.GetKey()
	if ok || k != "" {
		t.Errorf("GetKey on empty handler: got (%q, %v), want ('', false)", k, ok)
	}
}
