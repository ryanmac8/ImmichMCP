package models

import (
	"testing"
)

func TestSuccess(t *testing.T) {
	meta := NewMeta("http://immich.example.com")
	resp := Success("hello", meta)
	if !resp.Ok {
		t.Error("expected Ok=true")
	}
	if resp.Error != nil {
		t.Error("expected no error")
	}
	if resp.Result != "hello" {
		t.Errorf("unexpected result: %v", resp.Result)
	}
	if resp.Meta.ImmichBaseURL != "http://immich.example.com" {
		t.Error("unexpected base URL")
	}
	if resp.Meta.RequestID == "" {
		t.Error("expected non-empty request ID")
	}
}

func TestErrorResponse(t *testing.T) {
	meta := NewMeta("http://immich.example.com")
	resp := ErrorResponse(ErrNotFound, "not found", nil, meta)
	if resp.Ok {
		t.Error("expected Ok=false")
	}
	if resp.Error == nil {
		t.Fatal("expected error")
	}
	if resp.Error.Code != ErrNotFound {
		t.Errorf("unexpected code: %s", resp.Error.Code)
	}
	if resp.Error.Message != "not found" {
		t.Errorf("unexpected message: %s", resp.Error.Message)
	}
}

func TestNewMeta(t *testing.T) {
	m1 := NewMeta("http://a.example.com")
	m2 := NewMeta("http://a.example.com")
	if m1.RequestID == m2.RequestID {
		t.Error("expected unique request IDs")
	}
}
