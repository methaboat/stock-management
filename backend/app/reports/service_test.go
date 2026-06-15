package reports

import (
	"testing"
)

func TestCalculateValuation_Basic(t *testing.T) {
	baseCost := 10.0
	qtyOnHand := 15.0

	// Chronological layers (oldest to newest):
	layers := []CostLayer{
		{Quantity: 10, Cost: 12.0}, // Oldest
		{Quantity: 10, Cost: 14.0}, // Newest
	}

	fifo, lifo, avg := CalculateValuation(baseCost, qtyOnHand, layers)

	expectedFIFO := 200.0
	if fifo != expectedFIFO {
		t.Errorf("FIFO failed: got %v, want %v", fifo, expectedFIFO)
	}

	expectedLIFO := 190.0
	if lifo != expectedLIFO {
		t.Errorf("LIFO failed: got %v, want %v", lifo, expectedLIFO)
	}

	expectedAvg := 195.0
	if avg != expectedAvg {
		t.Errorf("Avg Cost failed: got %v, want %v", avg, expectedAvg)
	}
}

func TestCalculateValuation_Overflow(t *testing.T) {
	baseCost := 10.0
	qtyOnHand := 25.0

	layers := []CostLayer{
		{Quantity: 10, Cost: 12.0},
		{Quantity: 10, Cost: 14.0},
	}

	fifo, lifo, avg := CalculateValuation(baseCost, qtyOnHand, layers)

	expectedFIFO := 310.0
	if fifo != expectedFIFO {
		t.Errorf("FIFO overflow failed: got %v, want %v", fifo, expectedFIFO)
	}

	expectedLIFO := 310.0
	if lifo != expectedLIFO {
		t.Errorf("LIFO overflow failed: got %v, want %v", lifo, expectedLIFO)
	}

	expectedAvg := 325.0
	if avg != expectedAvg {
		t.Errorf("Avg Cost overflow failed: got %v, want %v", avg, expectedAvg)
	}
}

func TestCalculateValuation_ZeroStock(t *testing.T) {
	baseCost := 10.0
	qtyOnHand := 0.0

	layers := []CostLayer{
		{Quantity: 10, Cost: 12.0},
	}

	fifo, lifo, avg := CalculateValuation(baseCost, qtyOnHand, layers)

	if fifo != 0 || lifo != 0 || avg != 0 {
		t.Errorf("Zero stock failed: got fifo=%v, lifo=%v, avg=%v; want 0", fifo, lifo, avg)
	}
}

func TestCalculateValuation_NoLayers(t *testing.T) {
	baseCost := 10.0
	qtyOnHand := 5.0
	var layers []CostLayer

	fifo, lifo, avg := CalculateValuation(baseCost, qtyOnHand, layers)

	expected := 50.0
	if fifo != expected || lifo != expected || avg != expected {
		t.Errorf("No layers failed: got fifo=%v, lifo=%v, avg=%v; want %v", fifo, lifo, avg, expected)
	}
}
