package main

import (
	"log"
)

func main() {
	server := NewServer()
	// TODO **Multiple** logs of `use of closed network connection` error
	//		in stdout, meaning someone tries to use s.ln somewhere after shutdown

	// remove comment to test quit channel
	//go func() {
	//	time.Sleep(10 * time.Second)
	//	server.quitch <- struct{}{}
	//}()
	//------------------------------------

	log.Fatal(server.Start())

}
