// Package notifier is refactoring exercise 03: interfaces & abstraction (medium).
//
// GOAL
//
//	Decouple this notifier from the ambient world so it is easy to test and
//	reuse, without changing what a caller observes. `go test ./...` must stay
//	green throughout.
//
// SMELLS TO HUNT
//
//   - hidden dependencies reached for inside the method: os.Getenv for the
//     "from" address and time.Now for the timestamp. These make the output
//     nondeterministic and force tests to poke at global state.
//   - message formatting hard-wired inline, so it cannot be unit-tested on
//     its own and cannot be reused for other message types.
//   - validation done after the expensive work of building the body.
//   - the returned error is passed through bare, with no context.
//
// Suggested direction: move "from", the clock, and anything else ambient into
// explicit dependencies (constructor params or struct fields — a Clock can be
// a `func() time.Time`). Extract body formatting into a pure function you can
// test directly. Validate first. Wrap errors with context. The Sender
// interface is already here as the example to follow — do the same for the
// rest.
package notifier

import (
	"errors"
	"fmt"
	"time"
)

// Sender delivers a single message. This is the one dependency that is
// already abstracted — use it as the template for the others.
type Sender interface {
	Send(to, subject, body string) error
}

// Notifier sends order notifications.
type Notifier struct {
	fromEmail string
	sender    Sender
}

var ErrEmailRequired = errors.New("notifier: from email is required")
var ErrOrderIdRequired = errors.New("notifier: order ID is required")
var ErrSenderNeeded = errors.New("notifier: sender is required")
var ErrFromEmailNeeded = errors.New("notifier: from e-mail is required")

// Creates a new notifier with the provided sender and from e-mail.
func New(sender Sender, fromEmail string) (*Notifier, error) {
	if sender == nil {
		return nil, ErrSenderNeeded
	}

	if fromEmail == "" {
		return nil, ErrFromEmailNeeded
	}

	return &Notifier{sender: sender, fromEmail: fromEmail}, nil
}

func formatMessageBody(orderID, fromEmail string, shippedDate time.Time) string {
	return fmt.Sprintf("Hello,\n\n"+
		"Order %s shipped at %s.\n\n"+
		"Thanks,\n%s", orderID, shippedDate.Format(time.RFC3339), fromEmail)
}

func formatMessageSubject(orderID string) string {
	return fmt.Sprintf("Your order %s has shipped", orderID)
}

// NotifyShipment tells a customer their order has shipped.
func (n *Notifier) NotifyShipment(orderID, toEmail string, shippedDate time.Time) error {
	if toEmail == "" {
		return ErrEmailRequired
	}

	if orderID == "" {
		return ErrOrderIdRequired
	}

	subject := formatMessageSubject(orderID)
	body := formatMessageBody(orderID, n.fromEmail, shippedDate)

	if err := n.sender.Send(toEmail, subject, body); err != nil {
		return fmt.Errorf("failed to send message to %s: %w", toEmail, err)
	}
	return nil
}
