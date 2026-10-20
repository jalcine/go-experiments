package notifier

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// fakeSender records the last message it was asked to send.
type fakeSender struct {
	to, subject, body string
	called            int
	err               error
}

func (f *fakeSender) Send(to, subject, body string) error {
	f.called++
	f.to, f.subject, f.body = to, subject, body
	return f.err
}

func TestNotifyShipment(t *testing.T) {
	fs := &fakeSender{}
	datetime := time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)

	n, err := New(fs, "no-reply@service.net")
	if err != nil {
		t.Fatalf("failed to construct: %s", err)
	}

	if err := n.NotifyShipment("A1", "customer@example.com", datetime); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fs.called != 1 {
		t.Fatalf("Send called %d times, want 1", fs.called)
	}
	if fs.to != "customer@example.com" {
		t.Errorf("to = %q, want customer@example.com", fs.to)
	}
	// The order id must appear in both the subject and the body regardless of
	// how the message is formatted.
	if !strings.Contains(fs.subject, "A1") {
		t.Errorf("subject %q does not mention order A1", fs.subject)
	}
	if !strings.Contains(fs.body, "A1") {
		t.Errorf("body %q does not mention order A1", fs.body)
	}

	if !strings.Contains(fs.body, "no-reply@service.net") {
		t.Errorf("body %q does not mention from email", fs.body)
	}
	if !strings.Contains(fs.body, datetime.Format(time.RFC3339)) {
		t.Errorf("body %q does not mention the shipment date", fs.body)
	}
}

func TestNotifyShipmentRequiresEmail(t *testing.T) {
	fs := &fakeSender{}
	n, err := New(fs, "no-reply@service.net")
	if err != nil {
		t.Fatalf("failed to construct: %s", err)
	}

	if err := n.NotifyShipment("A1", "", time.Now()); err == nil {
		t.Error("expected error for empty email, got nil")
	}
	if fs.called != 0 {
		t.Errorf("Send called %d times for invalid input, want 0", fs.called)
	}
}

func TestNew(t *testing.T) {
	t.Run("missing sender", func(t *testing.T) {
		_, err := New(nil, "no-reply@service.net")
		if err == nil {
			t.Fatalf("no error was returned for missing sender")
		}

		if !errors.Is(err, ErrSenderNeeded) {
			t.Fatalf("lack of sender was not detected as an error")
		}

	})

	t.Run("missing from e-mail", func(t *testing.T) {
		_, err := New(&fakeSender{}, "")
		if err == nil {
			t.Fatalf("no error was returned for missing e-mail")
		}

		if !errors.Is(err, ErrFromEmailNeeded) {
			t.Fatalf("lack of from e-mail was not detected as an error")
		}
	})
}
