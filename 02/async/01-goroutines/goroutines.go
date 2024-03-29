package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

const (
	iterationsNum = 7
	goroutinesNum = 7
)

func doWork(th int) {
	for j := 0; j < iterationsNum; j++ {
		fmt.Printf(formatWork(th, j))
		// time.Sleep(time.Millisecond)
	}
}

func main() {
	for i := 0; i < goroutinesNum; i++ {
		go doWork(i)
	}
	fmt.Scanln()
}

func formatWork(in, j int) string {
	return fmt.Sprintln(strings.Repeat("  ", in), "█",
		strings.Repeat("  ", goroutinesNum-in),
		"th", in,
		"iter", j, strings.Repeat("■", j))
}

func imports() {
	fmt.Println(time.Millisecond, runtime.NumCPU())
}
