package entity

import "testing"

func TestDefaultChipSet(t *testing.T) {
	cs := DefaultChipSet()

	if len(cs.Denominations) != 6 {
		t.Fatalf("expected 6 denominations, got %d", len(cs.Denominations))
	}

	expectedValues := []int{1, 5, 10, 25, 100, 500}
	for i, d := range cs.Denominations {
		if d.Value != expectedValues[i] {
			t.Errorf("denomination %d: expected value %d, got %d", i, expectedValues[i], d.Value)
		}
		if d.Color == "" {
			t.Errorf("denomination %d: expected non-empty color", i)
		}
	}
}

func TestDefaultChipSetSorted(t *testing.T) {
	cs := DefaultChipSet()

	for i := 1; i < len(cs.Denominations); i++ {
		if cs.Denominations[i].Value <= cs.Denominations[i-1].Value {
			t.Errorf("denominations should be in ascending order: %d <= %d",
				cs.Denominations[i].Value, cs.Denominations[i-1].Value)
		}
	}
}

func TestChipSetCustom(t *testing.T) {
	cs := ChipSet{
		Denominations: []ChipDenomination{
			{Value: 10, Color: "#FF0000"},
			{Value: 50, Color: "#00FF00"},
		},
	}

	if len(cs.Denominations) != 2 {
		t.Fatalf("expected 2 denominations, got %d", len(cs.Denominations))
	}
}
