package main

import (
	"log"
)

func main() {
	server := NewServer()
	// remove comment to test quit channel
	//go func() {
	//	time.Sleep(10 * time.Second)
	//	server.quitch <- struct{}{}
	//}()
	//------------------------------------

	log.Fatal(server.Start())
}
