package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func main() {
	err := ioutil.WriteFile("./logs/logA.txt", []byte(strconv.Itoa(7)+"\n"), 0666)
	if err != nil {
		fmt.Println(err)
	}
	//Append second line
	file, err := os.OpenFile("./logs/logA.txt", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	if _, err := file.WriteString("second line"); err != nil {
		log.Fatal(err)
	}
	file, err = os.Open("./logs/logA.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	fmt.Println(scanner.Text() + " -> bakwas")
	for scanner.Scan() { // internally, it advances token based on sperator
		fmt.Println(scanner.Text()) // token in unicode-char
		//fmt.Println(scanner.Bytes()) // token in bytes

	}
}
