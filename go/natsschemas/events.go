// Package natsschemas defines the shared NATS message envelope and event
// payload types used across all services. The envelope MUST be used for every
// published message so consumers can rely on a consistent shape.
package natsschemas

import (
	"encoding/json"
	"time"
)

// EnvelopeVersion is the current shared envelope version. Bump only on
// breaking change to the envelope itself (not payloads).
const EnvelopeVersion = "1"

// Envelope wraps every NATS message published in the system.
type Envelope struct {
	Version    string          `json:"version"`
	ID         string          `json:"id"`
	OccurredAt time.Time       `json:"occurred_at"`
	TraceID    string          `json:"trace_id"`
	ActorID    string          `json:"actor_id,omitempty"`
	Payload    json.RawMessage `json:"payload"`
}

// Subjects enumerates every NATS subject in the system. Producers MUST use
// these constants — string literals at call sites are a code smell.
const (
	SubjectBookingCreated           = "booking.created"
	SubjectBookingConfirmed         = "booking.confirmed"
	SubjectBookingShifted           = "booking.shifted"
	SubjectBookingCancelled         = "booking.cancelled"
	SubjectBookingStarted           = "booking.started"
	SubjectBookingCompleted         = "booking.completed"
	SubjectBookingCapacityShortage  = "booking.capacity_shortage"
	SubjectBookingLockedAffected    = "booking.locked_affected"
	SubjectBookingAssignedBarber    = "booking.assigned_barber"

	SubjectQueueJoined           = "queue.joined"
	SubjectQueuePositionChanged  = "queue.position_changed"
	SubjectQueueYourTurn         = "queue.your_turn"

	SubjectBarberStatusChanged   = "barber.status_changed"
	SubjectBarberScheduleChanged = "barber.schedule_changed"

	SubjectPaymentSucceeded = "payment.succeeded"
	SubjectPaymentFailed    = "payment.failed"
	SubjectPaymentRefunded  = "payment.refunded"

	SubjectUserUpdated   = "user.updated"
	SubjectUserSuspended = "user.suspended"

	SubjectReviewCreated = "review.created"

	SubjectAuditLog = "audit.log"

	// Commands (request/reply).
	SubjectCmdNotificationSend = "notification.send"
	SubjectCmdAuthGetUser      = "auth.get_user"
)

// --- Event payloads ---

type BookingCreatedPayload struct {
	BookingID  string    `json:"booking_id"`
	CustomerID string    `json:"customer_id"`
	ShopID     string    `json:"shop_id"`
	StartTime  time.Time `json:"start_time"`
	IsLocked   bool      `json:"is_locked"`
}

type BookingConfirmedPayload struct {
	BookingID string `json:"booking_id"`
	PaymentID string `json:"payment_id"`
}

type BookingShiftedPayload struct {
	BookingID string    `json:"booking_id"`
	OldStart  time.Time `json:"old_start"`
	NewStart  time.Time `json:"new_start"`
	Reason    string    `json:"reason"`
}

type BookingCancelledPayload struct {
	BookingID    string `json:"booking_id"`
	CancelledBy  string `json:"cancelled_by"` // customer | barber | shop | system
	RefundStatus string `json:"refund_status,omitempty"`
}

type BookingCapacityShortagePayload struct {
	BookingID string    `json:"booking_id"`
	Deadline  time.Time `json:"deadline"`
}

type BookingLockedAffectedPayload struct {
	BookingID string `json:"booking_id"`
	ShopID    string `json:"shop_id"`
	Reason    string `json:"reason"`
}

type BookingAssignedBarberPayload struct {
	BookingID string `json:"booking_id"`
	BarberID  string `json:"barber_id"`
	Name      string `json:"name"`
}

type QueueJoinedPayload struct {
	QueueEntryID string `json:"queue_entry_id"`
	ShopID       string `json:"shop_id"`
	CustomerID   string `json:"customer_id"`
	Position     int    `json:"position"`
}

type QueuePositionChangedPayload struct {
	QueueEntryID string `json:"queue_entry_id"`
	Position     int    `json:"position"`
	EtaMinutes   int    `json:"eta_minutes"`
}

type QueueYourTurnPayload struct {
	QueueEntryID string `json:"queue_entry_id"`
	BarberName   string `json:"barber_name"`
}

type BarberStatusChangedPayload struct {
	BarberID string    `json:"barber_id"`
	Status   string    `json:"status"`
	FreeAt   time.Time `json:"free_at,omitempty"`
}

type BarberScheduleChangedPayload struct {
	BarberID string `json:"barber_id"`
	Reason   string `json:"reason"`
}

type PaymentSucceededPayload struct {
	PaymentIntentID string `json:"payment_intent_id"`
	BookingID       string `json:"booking_id"`
	AmountSatang    int64  `json:"amount_satang"`
}

type PaymentFailedPayload struct {
	PaymentIntentID string `json:"payment_intent_id"`
	BookingID       string `json:"booking_id"`
	FailureReason   string `json:"failure_reason"`
}

type PaymentRefundedPayload struct {
	PaymentIntentID string `json:"payment_intent_id"`
	BookingID       string `json:"booking_id"`
	AmountSatang    int64  `json:"amount_satang"`
}

type UserUpdatedPayload struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Language    string `json:"language"`
}

type UserSuspendedPayload struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason,omitempty"`
}

type ReviewCreatedPayload struct {
	ReviewID  string `json:"review_id"`
	BookingID string `json:"booking_id"`
	BarberID  string `json:"barber_id"`
	ShopID    string `json:"shop_id"`
	Stars     int    `json:"stars"`
}

type AuditLogPayload struct {
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Before     json.RawMessage `json:"before,omitempty"`
	After      json.RawMessage `json:"after,omitempty"`
}

// Encode wraps a payload in an Envelope and marshals to JSON.
func Encode(id, traceID, actorID string, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	env := Envelope{
		Version:    EnvelopeVersion,
		ID:         id,
		OccurredAt: time.Now().UTC(),
		TraceID:    traceID,
		ActorID:    actorID,
		Payload:    raw,
	}
	return json.Marshal(env)
}

// Decode parses an Envelope from a raw NATS message. Callers then
// json.Unmarshal env.Payload into the concrete payload type for the subject.
func Decode(b []byte) (Envelope, error) {
	var env Envelope
	err := json.Unmarshal(b, &env)
	return env, err
}
