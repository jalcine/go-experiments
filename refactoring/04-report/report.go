// Package report is refactoring exercise 04: structure & duplication (medium).
//
// GOAL
//
//	Break the one big function below into small, single-purpose pieces and
//	remove the copy-pasted blocks, without changing the report it produces.
//	`go test ./...` compares the exact output string, so it will tell you
//	immediately if behavior drifts.
//
// SMELLS TO HUNT
//
//   - a single function that parses, filters, aggregates, and formats — four
//     responsibilities that want to be four functions
//   - three near-identical blocks, one per region, that differ only in a
//     string literal (classic copy-paste; collapse them with a loop over a
//     map, and sort for deterministic output)
//   - magic strings for the region names repeated everywhere
//   - intermediate state (totals, counts) tracked in parallel variables that
//     a small struct or a map would model more clearly
//
// Suggested direction: a type for a Sale, a function to keep only completed
// sales, a function to total them by region, and a function to render the
// summary. GenerateReport then just wires those together.
package report

import (
	"fmt"
	"strings"
)

// Defines a region
type Area string

const (
	// Northen region.
	AreaNorth Area = "north"
	AreaSouth Area = "south"
	AreaEast  Area = "east"
	AreaWest  Area = "west"
)

// Sale is one row of raw sales data.
type Sale struct {
	Region    Area
	Amount    float64
	Completed bool
}

type Summary struct {
	Count int
	Total float64
}

type Summaries map[Area]Summary

var regions = []Area{AreaEast, AreaNorth, AreaSouth}

func Compose(sales []Sale) Summaries {
	summaries := map[Area]Summary{}

	for _, sale := range sales {
		if !sale.Completed {
			continue
		}

		summary := summaries[sale.Region]
		summaries[sale.Region] = Summary{
			Count: summary.Count + 1,
			Total: summary.Total + sale.Amount,
		}
	}

	return summaries
}

// Generate returns a human-readable summary of completed sales per
// region for the three regions the business operates in.
func Generate(sales []Sale) string {
	summaries := Compose(sales)
	grandTotal := 0.0

	var b strings.Builder
	b.WriteString("Sales Report\n")
	b.WriteString("============\n")
	for _, region := range regions {
		grandTotal += summaries[region].Total
		fmt.Fprintf(&b, "%s: %d sales, $%.2f\n", region, summaries[region].Count, summaries[region].Total)
	}
	fmt.Fprintf(&b, "TOTAL: $%.2f\n", grandTotal)
	return b.String()
}
