package main

import "fmt"
import "raft"

func main() {
	fmt.Println("Node starting...")
	c := raft.NewContainer("node-1")
	fmt.Println(c.Id())

}
