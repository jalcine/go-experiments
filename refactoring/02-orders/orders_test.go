package orders

import (
	"errors"
	"testing"
)

func TestNewOrder(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		o, err := NewOrder("A1", 2, 50)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if o.Status != OrderStatusPending {
			t.Errorf("new order status = %q, want pending", o.Status)
		}
	})

	invalid := map[string]struct {
		id      string
		items   int
		total   float64
		wantErr error
	}{
		"empty id":       {"", 1, 10, ErrNoOrderId},
		"zero items":     {"A1", 0, 10, ErrInvalidItemCount},
		"negative total": {"A1", 1, -1, ErrInvalidTotal},
	}
	for name, tc := range invalid {
		t.Run(name, func(t *testing.T) {
			_, err := NewOrder(tc.id, tc.items, tc.total)

			if err == nil {
				t.Fatalf("NewOrder(%q, %d, %v) = nil error, want error", tc.id, tc.items, tc.total)
			}

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("want %s; got %s", tc.wantErr, err)
			}
		})
	}
}

func TestApplyCoupon(t *testing.T) {
	t.Run("SAVE10", func(t *testing.T) {
		o, _ := NewOrder("A1", 1, 100)
		if err := o.ApplyCoupon(CouponSave10); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if o.Total != 90 {
			t.Errorf("total = %v, want 90", o.Total)
		}
	})
	t.Run("HALF over threshold", func(t *testing.T) {
		o, _ := NewOrder("A1", 1, 200)
		if err := o.ApplyCoupon(CouponHalf); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if o.Total != 100 {
			t.Errorf("total = %v, want 100", o.Total)
		}
	})
	t.Run("HALF under threshold", func(t *testing.T) {
		o, _ := NewOrder("A1", 1, 50)
		if err := o.ApplyCoupon(CouponHalf); err == nil {
			t.Error("expected error for HALF under threshold, got nil")
		}
	})
	t.Run("unknown code", func(t *testing.T) {
		o, _ := NewOrder("A1", 1, 50)
		if err := o.ApplyCoupon(33); err == nil {
			t.Error("expected error for unknown coupon, got nil")
		}
	})
}

func TestMarkPaid(t *testing.T) {
	o, _ := NewOrder("A1", 1, 50)
	if err := o.MarkPaid(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Status != OrderStatusPaid {
		t.Errorf("status = %d, want paid", o.Status)
	}
	if err := o.MarkPaid(); err == nil {
		t.Error("expected error paying an already-paid order, got nil")
	}
}

func TestAddItemsFromString(t *testing.T) {
	o, _ := NewOrder("A1", 1, 50)
	o.AddItemsFromString("3")
	if o.Items != 4 {
		t.Errorf("items = %d, want 4", o.Items)
	}
	// Malformed input is currently treated as zero.
	o.AddItemsFromString("not-a-number")
	if o.Items != 4 {
		t.Errorf("items = %d, want 4 after bad input", o.Items)
	}
}
