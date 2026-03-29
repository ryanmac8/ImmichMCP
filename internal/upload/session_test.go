package upload

import (
	"testing"
	"time"
)

func TestCreateAndGetSession(t *testing.T) {
	svc := NewSessionService(time.Minute)
	defer svc.Close()

	name := "photo.jpg"
	fav := true
	sess := svc.CreateSession(&name, &fav, nil)

	if sess.SessionID == "" {
		t.Error("expected non-empty session ID")
	}
	if sess.Status != StatusPending {
		t.Errorf("expected pending, got %s", sess.Status)
	}
	if *sess.FileName != "photo.jpg" {
		t.Error("unexpected filename")
	}

	got := svc.GetSession(sess.SessionID)
	if got == nil {
		t.Fatal("expected session, got nil")
	}
	if got.SessionID != sess.SessionID {
		t.Error("session ID mismatch")
	}
}

func TestGetSessionNotFound(t *testing.T) {
	svc := NewSessionService(time.Minute)
	defer svc.Close()

	if svc.GetSession("nonexistent") != nil {
		t.Error("expected nil for unknown session")
	}
}

func TestSetCompleted(t *testing.T) {
	svc := NewSessionService(time.Minute)
	defer svc.Close()

	sess := svc.CreateSession(nil, nil, nil)
	svc.SetCompleted(sess.SessionID, "asset-123")

	got := svc.GetSession(sess.SessionID)
	if got.Status != StatusCompleted {
		t.Errorf("expected completed, got %s", got.Status)
	}
	if got.AssetID == nil || *got.AssetID != "asset-123" {
		t.Error("unexpected asset ID")
	}
}

func TestSetFailed(t *testing.T) {
	svc := NewSessionService(time.Minute)
	defer svc.Close()

	sess := svc.CreateSession(nil, nil, nil)
	svc.SetFailed(sess.SessionID, "upload error")

	got := svc.GetSession(sess.SessionID)
	if got.Status != StatusFailed {
		t.Errorf("expected failed, got %s", got.Status)
	}
	if got.Error == nil || *got.Error != "upload error" {
		t.Error("unexpected error message")
	}
}

func TestSessionExpiry(t *testing.T) {
	svc := NewSessionService(time.Millisecond)
	defer svc.Close()

	sess := svc.CreateSession(nil, nil, nil)
	time.Sleep(5 * time.Millisecond)

	got := svc.GetSession(sess.SessionID)
	if got == nil {
		t.Fatal("expected session to still exist (cleanup runs on ticker)")
	}
	if got.Status != StatusExpired {
		t.Errorf("expected expired, got %s", got.Status)
	}
}

func TestStatusString(t *testing.T) {
	cases := []struct {
		s    Status
		want string
	}{
		{StatusPending, "pending"},
		{StatusUploading, "uploading"},
		{StatusCompleted, "completed"},
		{StatusFailed, "failed"},
		{StatusExpired, "expired"},
		{Status(99), "unknown"},
	}
	for _, tc := range cases {
		if tc.s.String() != tc.want {
			t.Errorf("Status(%d).String() = %q, want %q", tc.s, tc.s.String(), tc.want)
		}
	}
}
