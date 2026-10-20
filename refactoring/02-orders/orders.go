// Package orders is refactoring exercise 02: error handling (easy-medium).
//
// GOAL
//
//	Make errors first-class and idiomatic without changing the observable
//	contract: the same inputs must still succeed or fail the same way, so
//	`go test ./...` stays green throughout.
//
// SMELLS TO HUNT
//
//   - errors built ad hoc with errors.New(fmt.Sprintf(...)); no sentinels,
//     no wrapping, so callers cannot use errors.Is / errors.As
//   - a "status" modelled as a bare string that any typo can corrupt,
//     instead of a defined, closed set of values
//   - magic numbers (discount rates, thresholds) sprinkled inline
//   - deep nesting where early returns would read better
//   - a swallowed error (look for `_ =`) that hides bad input
//
// Suggested direction: introduce a Status type with constants, package-level
// sentinel errors (var ErrXxx = errors.New(...)), wrap with fmt.Errorf("...: %w"),
// name the constants, and flatten the happy path. Add errors.Is-based
// assertions of your own once the sentinels exist — the existing tests will
// keep passing.
package orders

import (
	"errors"
	"fmt"
	"strconv"
)

type OrderStatus int

const (
	OrderStatusPending OrderStatus = iota
	OrderStatusPaid
)

func (s OrderStatus) String() string {
	switch s {
	case OrderStatusPaid:
		return "paid"
	case OrderStatusPending:
		return "pending"
	default:
		return fmt.Sprintf("unknown OrderStatus(%d)", int(s))
	}
}

// Order is a customer order moving through a tiny lifecycle.
type Order struct {
	ID     string
	Items  int
	Total  float64
	Status OrderStatus
}

var (
	ErrInvalidItemCount           = errors.New("count of items must be positive")
	ErrInvalidTotal               = errors.New("total must not be negative")
	ErrNoOrderId                  = errors.New("no order ID was provided")
	ErrCouponNotApplicable        = errors.New("coupon condition was not met")
	ErrCouponNotFound             = errors.New("coupon not found")
	ErrOrderInvalidStateForPaying = errors.New("cannot pay order if not in pending state")
)

// NewOrder validates and constructs an order in its initial state.
func NewOrder(id string, items int, total float64) (*Order, error) {
	if id == "" {
		return nil, ErrNoOrderId
	}

	if items <= 0 {
		return nil, fmt.Errorf("order #%s has %d items: %w", id, items, ErrInvalidItemCount)
	}

	if total < 0 {
		return nil, fmt.Errorf("order #%s has total %.2f: %w", id, total, ErrInvalidTotal)
	}

	return &Order{ID: id, Items: items, Total: total, Status: OrderStatusPending}, nil
}

type Coupon int

const (
	CouponHalf Coupon = iota
	CouponSave10
)

func (s Coupon) String() string {
	switch s {
	case CouponHalf:
		return "HALF"
	case CouponSave10:
		return "SAVE10"
	default:
		return fmt.Sprintf("Coupon(%d)", s)
	}
}

// ApplyCoupon reduces the order total according to a coupon code. Known codes
// are "SAVE10" (10% off) and "HALF" (50% off, but only on totals over 100).
func (o *Order) ApplyCoupon(code Coupon) error {
	switch code {
	case CouponSave10:
		o.Total = o.Total * 0.9
	case CouponHalf:
		if o.Total < 100 {
			return fmt.Errorf("total: %.2f, minimum needed is 100 for HALF coupon: %w", o.Total, ErrCouponNotApplicable)
		}
		o.Total = o.Total * 0.5
	default:
		return fmt.Errorf("order %s: coupon %s: %w", o.ID, code, ErrCouponNotFound)
	}

	return nil
}

// MarkPaid transitions the order to the paid state. Only a pending order may
// be paid.
func (o *Order) MarkPaid() error {
	if o.Status == OrderStatusPending {
		o.Status = OrderStatusPaid
		return nil
	}
	return fmt.Errorf("order %s: status %s: %w", o.ID, o.Status, ErrOrderInvalidStateForPaying)
}

// AddItemsFromString parses a quantity from user input and adds it to the
// order
func (o *Order) AddItemsFromString(qty string) error {
	n, err := strconv.Atoi(qty)
	if err != nil {
		return fmt.Errorf("got %s as a quantity, expected an integer: %w", qty, err)
	}
	o.Items = o.Items + n
	return nil
}
