package main

import (
	// "fmt"

	"github.com/eniimz/cas/p2p"
)

func main() {
	tr := p2p.NewTCPTransport(":4000")

	tr.ListenAndAccept()

	// for msg := range tr.Consume() {
	// 	fmt.Println("The message: %v\n", msg)
	// }

	select {}
}
