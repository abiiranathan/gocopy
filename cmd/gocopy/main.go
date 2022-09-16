package main

import (
	"fmt"
	"log"
	"os"
	"time"

	gocopy "github.com/abiiranathan/gocopy/pkg"
)

// could accept flags
func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s [SOURCE] [DESTINATION]\n", os.Args[0])
	}

	start := time.Now()
	copier := gocopy.New(gocopy.Verbose())
	copier.CopyDir(os.Args[1], os.Args[2])

	took := time.Since(start)
	printDuration(took)

}

func printDuration(took time.Duration) {
	ms := took.Milliseconds()
	SEC := 1000
	MIN := SEC * 60
	HOUR := MIN * 60

	if ms >= (int64(HOUR)) {
		fmt.Printf("\nTook: %.2f hours\n", took.Hours())
	} else if ms >= (int64(MIN)) {
		fmt.Printf("\nTook: %.2f min\n", took.Minutes())
	} else if ms >= (int64(SEC)) {
		fmt.Printf("\nTook: %.2f sec\n", took.Seconds())
	} else {
		fmt.Printf("\nTook: %d msec\n", ms)
	}
}
