package main

import (
	"../../quiz05"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
	// "strings"
)

func main() {
	args := os.Args[1:]
	//args[0] = name
	//args[1] = Listening Port
	//args[2] = Dialling Port 1
	//args[3] = Dialling Port 2
	//.
	//.
	//.
	//args[x] = Dialling Port (x-1)
	if len(args) < 3 {
		log.Fatalln("Error 1: Program needs at least three valid arguments. Only " + strconv.Itoa(len(args)) + " provided")
	}
	rand.Seed(time.Now().UnixNano())
	myNode := quiz05.Node{args[0], "localhost", args[1], args[2:], "xxxx", quiz05.Follower, 0, nil,
		nil, -1, -1, nil, nil}
	fmt.Println("Successfully created node")
	go myNode.Wait()
	myNode.NodeExecute()

}
