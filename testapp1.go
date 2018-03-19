package main

import (
	"./movement"
	"fmt"
)

func main() {

	m, err := movement.CreateLocationGraph("movement/example_file")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Note: this file should be run from the root directory of the project")
		return
	}

	fn := func(p movement.Point) {
		fmt.Printf("%.2f %.2f\n", p.X, p.Y)
	}

	movement.Travel(movement.Point{0.0, 0.0}, m, fn)
}
