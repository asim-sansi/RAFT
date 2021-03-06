package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

type Protocol struct {
	Request   []string //name of request. Example: RTC => Request to Connect
	Operation string
	Value     int
	Port      string
	Name      string
}

func ConnnectToNode(Port string, IP string) net.Conn {
	//dial connection to a certain node
	conn, err := net.Dial("tcp", IP+":"+Port)
	if err != nil {
		fmt.Println("Error: Unable to connect to the specified target...")
		return nil
		// handle error
	}
	return conn
}

// Wait : waits for further contact
func Wait(Port string) {
	ln, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		log.Fatal(err)
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}

		handleConnection(conn)
	}
}

func handleConnection(c net.Conn) {
	message := make([]byte, 150)
	l, err := c.Read(message)
	if err != nil {
		fmt.Println("error at receiving message")
		return
	}
	fmt.Println()
	fmt.Println(string(message)[:l])

}

func main() {

	args := os.Args[1:]
	//args[0] = name
	//args[1] = Listening Port
	if len(args) > 3 {
		log.Fatalln("Error 1: Program needs at most three valid arguments. " + strconv.Itoa(len(args)) + " provided")
	}
	go Wait(args[1])

	for {
		fmt.Println("Please Choose From The Options Below")
		fmt.Println("1: Connect to a server")
		fmt.Println("2: Exit")
		var Input string
		fmt.Scanln(&Input)
		switch Input {
		case "1":
			fmt.Print("    Enter port -> ")
			var port string
			fmt.Scanln(&port)
			var operation string
			var value int
			for operation != "EXIT" {
				operation = "INVALID"
				for operation == "INVALID" {
					fmt.Println("    Choose An Operation By Entering Corresponding Menu Number")
					fmt.Println("    	1. Addition[+]")
					fmt.Println("    	2. Subtraction[-]")
					fmt.Println("    	3. Multiplication[*]")
					fmt.Println("    	4. Assignment[=]")
					fmt.Println("    	5. Get State Machine Value")
					fmt.Println("    	6. Exit")
					fmt.Print("    Enter Operation -> ")
					fmt.Scanln(&operation)
					switch operation {
					case "1":
						operation = "+"
						break
					case "2":
						operation = "-"
						break
					case "3":
						operation = "*"
						break
					case "4":
						operation = "="
						break
					case "5":
						operation = "GET"
						break
					case "6":
						operation = "EXIT"
						break
					default:
						operation = "INVALID"
						fmt.Println("        *Invalid Operation, Choose Again")
					}
				}
				if operation == "6" {
					break
				}
				if operation != "5" {
					fmt.Print("    Enter Value -> ")
					fmt.Scanln(&value)
				}

				c := ConnnectToNode(port, "localhost")
				if c == nil {
					break
				}
				var datapacket Protocol
				datapacket.Request = []string{"COMMAND"}
				datapacket.Name = args[0]
				datapacket.Port = args[1]
				datapacket.Operation = operation
				datapacket.Value = value
				var byteSack bytes.Buffer
				enc := gob.NewEncoder(&byteSack)
				_ = enc.Encode(datapacket)
				c.Write(byteSack.Bytes())
				c.Close()
			}
		case "2":
			fmt.Println("Exiting")
			return
		default:
			fmt.Println("Invalid Choice")
		}
	}
}
