package divination

import (
	"testing"
)

func TestLinesToBinary(t *testing.T) {
	lines := []modelLine{
		{IsYang: 1}, {IsYang: 1}, {IsYang: 1}, {IsYang: 1}, {IsYang: 1}, {IsYang: 1},
	}
	if got := linesToBinaryFromTest(lines); got != "111111" {
		t.Fatalf("expected 111111, got %s", got)
	}
}

func TestChangedBinary(t *testing.T) {
	lines := []modelLine{
		{Position: 1, IsYang: 1, IsMoving: 0},
		{Position: 2, IsYang: 1, IsMoving: 0},
		{Position: 3, IsYang: 1, IsMoving: 0},
		{Position: 4, IsYang: 1, IsMoving: 0},
		{Position: 5, IsYang: 0, IsMoving: 0},
		{Position: 6, IsYang: 1, IsMoving: 1},
	}
	if got := changedBinaryFromTest(lines); got != "111100" {
		t.Fatalf("expected 111100, got %s", got)
	}
}

type modelLine struct {
	Position  int
	IsYang    int
	IsMoving  int
}

func linesToBinaryFromTest(lines []modelLine) string {
	b := make([]byte, 6)
	for i, line := range lines {
		if line.IsYang == 1 {
			b[i] = '1'
		} else {
			b[i] = '0'
		}
	}
	return string(b)
}

func changedBinaryFromTest(lines []modelLine) string {
	b := make([]byte, 6)
	for i, line := range lines {
		yang := line.IsYang == 1
		if line.IsMoving == 1 {
			yang = !yang
		}
		if yang {
			b[i] = '1'
		} else {
			b[i] = '0'
		}
	}
	return string(b)
}
