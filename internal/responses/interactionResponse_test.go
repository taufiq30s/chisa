package responses

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

// An empty ActionsRow makes Discord reject the whole response ("Invalid Form
// Body"), which surfaces to the user as a permanently deferred /music list.
// Calling AddNewRow() before any button must stay a no-op, and the buttons
// added afterwards must still land in a single row.
func TestAddNewRowSkipsEmptyRows(t *testing.T) {
	r := InteractionResponse(nil, &discordgo.Interaction{})

	r.AddNewRow() // no pending components yet
	if len(r.components) != 0 {
		t.Fatalf("AddNewRow with no pending components appended a row: %d", len(r.components))
	}

	r.WithButton("Previous", discordgo.PrimaryButton, "queue-previous-1", true).
		WithButton("Next", discordgo.PrimaryButton, "queue-next-1", false).
		AddNewRow()

	if len(r.components) != 1 {
		t.Fatalf("expected 1 component row, got %d", len(r.components))
	}
	row, ok := r.components[0].(discordgo.ActionsRow)
	if !ok {
		t.Fatalf("expected discordgo.ActionsRow, got %T", r.components[0])
	}
	if len(row.Components) != 2 {
		t.Fatalf("expected 2 buttons in the row, got %d", len(row.Components))
	}

	// A second empty flush must not append another row.
	r.AddNewRow()
	if len(r.components) != 1 {
		t.Fatalf("AddNewRow with no pending components appended a row: %d", len(r.components))
	}
}
