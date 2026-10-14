package main

import (
	"fmt"
	"math"
)

const s string = "constant value"

func main() {
	fmt.Println(s)

	const n = 5_000_000_000

	const d = 3e20 / n

	fmt.Println(d)
	fmt.Println(int64(d))

	fmt.Println(math.Sin(n))
}
