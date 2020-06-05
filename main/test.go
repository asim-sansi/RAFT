package main

import (
	"../../raft"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {

	file, err := os.Open("./logs/bakwas.txt")
	if err != nil {
		fmt.Println(err)
		file, err = os.Create("./logs/bakwas.txt")
		if err != nil {
			fmt.Println(err)
		}
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
			fmt.Println(len(entry))
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
	fmt.Println(FileTerm)
	fmt.Println(len(FileLog))
}
