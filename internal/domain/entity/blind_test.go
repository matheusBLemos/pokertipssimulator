package entity

import "testing"

func TestBlindStructureDefaults(t *testing.T) {
	bs := BlindStructure{
		Levels: []BlindLevel{
			{SmallBlind: 5, BigBlind: 10, Ante: 0, Duration: 600},
			{SmallBlind: 10, BigBlind: 20, Ante: 0, Duration: 600},
			{SmallBlind: 25, BigBlind: 50, Ante: 5, Duration: 600},
		},
		CurrentLevel: 0,
	}

	if bs.Levels[0].SmallBlind != 5 {
		t.Errorf("expected SB 5, got %d", bs.Levels[0].SmallBlind)
	}
	if bs.Levels[0].BigBlind != 10 {
		t.Errorf("expected BB 10, got %d", bs.Levels[0].BigBlind)
	}
	if len(bs.Levels) != 3 {
		t.Errorf("expected 3 levels, got %d", len(bs.Levels))
	}
}

func TestBlindLevelWithAnte(t *testing.T) {
	level := BlindLevel{SmallBlind: 25, BigBlind: 50, Ante: 5, Duration: 600}
	if level.Ante != 5 {
		t.Errorf("expected ante 5, got %d", level.Ante)
	}
}

func TestBlindLevelManualDuration(t *testing.T) {
	level := BlindLevel{SmallBlind: 10, BigBlind: 20, Duration: 0}
	if level.Duration != 0 {
		t.Error("expected Duration 0 for manual advancement")
	}
}

func TestBlindStructureLevelProgression(t *testing.T) {
	bs := BlindStructure{
		Levels: []BlindLevel{
			{SmallBlind: 5, BigBlind: 10},
			{SmallBlind: 10, BigBlind: 20},
			{SmallBlind: 25, BigBlind: 50},
		},
		CurrentLevel: 0,
	}

	for i := 0; i < len(bs.Levels)-1; i++ {
		bs.CurrentLevel++
	}

	if bs.CurrentLevel != 2 {
		t.Errorf("expected current level 2, got %d", bs.CurrentLevel)
	}

	current := bs.Levels[bs.CurrentLevel]
	if current.SmallBlind != 25 || current.BigBlind != 50 {
		t.Errorf("expected 25/50, got %d/%d", current.SmallBlind, current.BigBlind)
	}
}
