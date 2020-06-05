package main

import (
	"../../raft"
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
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
	//Initializing Log from Log File
	fmt.Println("Initializing Log from Log File")
	file, err := os.Open("../logs/log" + args[0] + ".txt")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Log File unreachable...")
		return
	}
	defer file.Close()

	var FileLog []raft.LogEntry
	FileTerm := 0
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	if scanner.Text() != "" {
		value, err := strconv.Atoi(scanner.Text())
		if err != nil {
			fmt.Println("Error at Parsing Log Term Value  To Integer")
			return
		}
		FileTerm = value
	}
	for scanner.Scan() { // internally, it advances token based on sperator
		entry := strings.Split(scanner.Text(), ",")
		if len(entry) == 3 {
			//fmt.Println(len(entry))
			EntryTerm, err := strconv.Atoi(entry[0])
			if err != nil {
				fmt.Println("Error at Parsing Log Entry Term Value To Integer")
				return
			}
			EntryValue, err := strconv.Atoi(entry[2])
			if err != nil {
				fmt.Println("Error at Parsing Log Entry Delta To Integer")
				return
			}
			FileLog = append(FileLog, raft.LogEntry{Term: EntryTerm, Operation: entry[1], Value: EntryValue})
		} else {
			fmt.Println("Invalid Log Entry In Log File, quitting [Check The Logs]")
			return
		}
	}

	rand.Seed(time.Now().UnixNano())
	myNode := raft.Node{args[0], "localhost", args[1], args[2:], "xxxx", raft.Follower, FileTerm, nil,
		FileLog, -1, -1, nil, nil}
	fmt.Println("Successfully created node with Log ")
	fmt.Println(FileLog)
	go myNode.Wait()
	myNode.NodeExecute()

}
