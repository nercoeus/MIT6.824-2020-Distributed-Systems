package main

//
// start the master process, which is implemented
// in ../mr/master.go
//
// go run mrmaster.go pg*.txt
//
// Please do not change this file.
//

import (
	"../mr"
)
import "time"
import "os"
import "fmt"

func main() {
	var filename []string
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: mrmaster inputfiles...\n")
		filename = append(filename, "pg-being_ernest.txt", "pg-dorian_gray.txt", "pg-frankenstein.txt", "pg-grimm.txt")
		filename = append(filename, "pg-huckleberry_finn.txt", "pg-metamorphosis.txt", "pg-sherlock_holmes.txt", "pg-tom_sawyer.txt")
	} else {
		filename = os.Args[1:]
	}
	m := mr.MakeMaster(filename, 10)
	for m.Done() == false {
		time.Sleep(time.Second)
	}

	time.Sleep(time.Second)
}
