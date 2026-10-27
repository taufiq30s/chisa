package moderation

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

func TestVerificationButtonsCarryMemberIDAndDisabledState(t *testing.T) {
	for _, disabled := range []bool{false, true} {
		row := verificationButtons("42", disabled)[0].(discordgo.ActionsRow)
		accept, reject := row.Components[0].(discordgo.Button), row.Components[1].(discordgo.Button)

		if accept.CustomID != "acc-req-accept-42" || reject.CustomID != "acc-req-reject-42" {
			t.Fatalf("unexpected custom ids: %q, %q", accept.CustomID, reject.CustomID)
		}
		if accept.Disabled != disabled || reject.Disabled != disabled {
			t.Fatalf("disabled state not applied for disabled=%v", disabled)
		}
	}
}

func TestLockVerificationSerialisesSameMember(t *testing.T) {
	unlock := lockVerification("42")
	acquired := make(chan struct{})
	go func() {
		unlockSecond := lockVerification("42")
		close(acquired)
		unlockSecond()
	}()

	select {
	case <-acquired:
		t.Fatal("same member acquired the lock concurrently")
	case <-time.After(50 * time.Millisecond):
	}

	unlock()
	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("lock was never released")
	}
}

func TestLockVerificationAllowsDifferentMembers(t *testing.T) {
	unlockA := lockVerification("a")
	defer unlockA()

	done := make(chan struct{})
	go func() {
		unlockB := lockVerification("b")
		unlockB()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("different members blocked each other")
	}
}

func TestPendingDecisionIsNoOpWithoutRedis(t *testing.T) {
	redisClient = nil

	savePendingDecision(verificationDecision{Action: verifyActionAccept, MemberID: "42"})
	clearPendingDecision("42")
	ReplayPendingVerifications(nil)
}
