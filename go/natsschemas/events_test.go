package natsschemas

import (
	"encoding/json"
	"testing"
)

func TestEncodeDecodeRoundtrip(t *testing.T) {
	in := BookingCreatedPayload{
		BookingID:  "bkg_1",
		CustomerID: "usr_1",
		ShopID:     "shp_1",
		IsLocked:   false,
	}
	b, err := Encode("msg_1", "trace_1", "usr_actor", in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	env, err := Decode(b)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Version != EnvelopeVersion {
		t.Fatalf("version=%q want=%q", env.Version, EnvelopeVersion)
	}
	if env.ID != "msg_1" || env.TraceID != "trace_1" || env.ActorID != "usr_actor" {
		t.Fatalf("envelope metadata mismatch: %+v", env)
	}

	var out BookingCreatedPayload
	if err := json.Unmarshal(env.Payload, &out); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if out != in {
		t.Fatalf("payload roundtrip mismatch: got=%+v want=%+v", out, in)
	}
}

func TestSubjectsAreUnique(t *testing.T) {
	all := []string{
		SubjectBookingCreated, SubjectBookingConfirmed, SubjectBookingShifted,
		SubjectBookingCancelled, SubjectBookingStarted, SubjectBookingCompleted,
		SubjectBookingCapacityShortage, SubjectBookingLockedAffected,
		SubjectBookingAssignedBarber,
		SubjectQueueJoined, SubjectQueuePositionChanged, SubjectQueueYourTurn,
		SubjectBarberStatusChanged, SubjectBarberScheduleChanged,
		SubjectPaymentSucceeded, SubjectPaymentFailed, SubjectPaymentRefunded,
		SubjectUserUpdated, SubjectUserSuspended,
		SubjectReviewCreated, SubjectAuditLog,
		SubjectCmdNotificationSend, SubjectCmdAuthGetUser,
	}
	seen := make(map[string]struct{}, len(all))
	for _, s := range all {
		if _, dup := seen[s]; dup {
			t.Fatalf("duplicate subject: %q", s)
		}
		seen[s] = struct{}{}
	}
}
