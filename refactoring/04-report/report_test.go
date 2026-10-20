package report

import "testing"

func TestGenerate(t *testing.T) {
	sales := []Sale{
		{Region: AreaNorth, Amount: 100, Completed: true},
		{Region: AreaNorth, Amount: 50, Completed: false}, // dropped: not completed
		{Region: AreaSouth, Amount: 200, Completed: true},
		{Region: AreaSouth, Amount: 25, Completed: true},
		{Region: AreaEast, Amount: 75, Completed: true},
		{Region: AreaWest, Amount: 999, Completed: true}, // dropped: region not reported
	}

	want := "Sales Report\n" +
		"============\n" +
		"east: 1 sales, $75.00\n" +
		"north: 1 sales, $100.00\n" +
		"south: 2 sales, $225.00\n" +
		"TOTAL: $400.00\n"

	if got := Generate(sales); got != want {
		t.Errorf("Generate() mismatch\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestGenerateReportEmpty(t *testing.T) {
	want := "Sales Report\n" +
		"============\n" +
		"east: 0 sales, $0.00\n" +
		"north: 0 sales, $0.00\n" +
		"south: 0 sales, $0.00\n" +
		"TOTAL: $0.00\n"

	if got := Generate(nil); got != want {
		t.Errorf("Generate(nil) mismatch\n got:\n%s\nwant:\n%s", got, want)
	}
}
