package entity

import "testing"

func TestPotCreation(t *testing.T) {
	pot := Pot{
		Amount:      100,
		EligibleIDs: []string{"p1", "p2"},
	}

	if pot.Amount != 100 {
		t.Errorf("expected amount 100, got %d", pot.Amount)
	}
	if len(pot.EligibleIDs) != 2 {
		t.Fatalf("expected 2 eligible players, got %d", len(pot.EligibleIDs))
	}
}

func TestPotZeroAmount(t *testing.T) {
	pot := Pot{Amount: 0}
	if pot.Amount != 0 {
		t.Errorf("expected 0, got %d", pot.Amount)
	}
	if pot.EligibleIDs != nil {
		t.Error("expected nil eligible IDs for empty pot")
	}
}
