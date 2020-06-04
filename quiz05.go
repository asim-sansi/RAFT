package quiz05

import (
	// "bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"

	// "os"
	"strconv"
	// "strings"
	"time"
)

//Variable to be in Sync
var Variable int

// HeartbeatInterval is the interval after which leader sends heartbeat ping to its followers
var HeartbeatInterval = 3
var mychannel chan string

// Role : For nodes
type Role int

// Roles Defined
const (
	Follower  = iota
	Candidate = iota
	Leader    = iota
)

//Log Structure
type LogEntry struct {
	Term      int
	Operation string
	Value     int
}

// Node : A node in the RAFT network
type Node struct {
	Name          string          //public key of node
	Server        string          //server node runs on
	Port          string          //port node runs on
	Myconnections []string        //details of nodes directly connected to
	LeaderPort    string          //holds port of current leader
	MyRole        Role            //current role
	Term          int             //current term
	Votes         map[string]bool //a list of votes received
	Log           []LogEntry      //Log entries to store (empty for heartbeat; may send more than one for efficiency)
	CommitIndex   int             //index of highest log entry known to be committed (initialized to 0, increasesmonotonically)
	LastApplied   int             //index of highest log entry applied to state machine (initialized to 0, increases monotonically)

	//Only for Leaders
	NextIndex  map[string]int //for each server, index of the next log entry to send to that server (initialized to leader last log index + 1)
	MatchIndex map[string]int //for each server, index of highest log entry known to be replicated on server (initialized to 0, increases monotonically)
}

// Protocol : Protocol for sent messages
type Protocol struct {
	Request []string //name of request. Example: RTC => Request to Connect
	Name    string   //name of node sent in a specific kind of request
	Server  string   //server of node sent in a specific kind of request
	Port    string   //port of node sent in a specific kind of request
	Term    int      //current term

	//For elections to check if the candidate's log is at least as updated as the voter
	LastLogIndex int
	LastLogTerm  int

	PrevLogIndex int        //index of log entry immediately preceding new ones
	LeaderCommit int        //Leaser's commit index
	PrevLogTerm  int        //term of prevLogIndex entry
	Entries      []LogEntry //Log entries to store (empty for heartbeat; may send more than one for efficiency)
	Operation    string
	Value        int
}

// ConnnectToNode : connects to a node
func (n *Node) ConnnectToNode(Port string) net.Conn {
	//dial connection to a certain node
	conn, err := net.Dial("tcp", n.Server+":"+Port)
	if err != nil {
		fmt.Println("Error: Unable to connect to the specified target...")
		return nil
		// handle error
	}
	return conn
}

// NodeExecute : waits for user input
func (n *Node) NodeExecute() {
	//execution starts by seeding random function
	rand.Seed(time.Now().UnixNano())

	//starting node main loop
	for true {
		switch n.MyRole { //executes current role of node. Initial is always follower
		case Follower:
			fmt.Println("I am Follower " + n.Name + " Now")
			//setting timeout to random number between 2*HeartBeatInterval to 3*HeartbeatInterval
			randomDuration := time.Duration(rand.Intn(HeartbeatInterval) + 2*HeartbeatInterval)

			//go routine runs in thread, and accepts connection requests
			mychannel = make(chan string, 1)

			select {
			//case connection is received
			case <-mychannel:
				// fmt.Println("Ping Received")
				break
			//case timeout occurs before a connection is received
			case <-time.After(randomDuration * time.Second):
				fmt.Println("Heartbeat Time out As Follower")
				//node role automatically turns to Candidate
				n.MyRole = Candidate
			}
		case Candidate:
			fmt.Println("I am Candidate " + n.Name + " Now")
			//candidate node automatically updates its term
			n.Term++
			//node intializes vote map for collection of votes
			n.Votes = make(map[string]bool)
			//giving vote to self
			n.Votes[n.Port] = true
			//sets leaderport to null
			n.LeaderPort = "xxxx"
			//constructs new vote request
			var dataPacket Protocol // buffer that holds converted bytes
			dataPacket.Request = []string{"VOTE"}
			dataPacket.Term = n.Term
			dataPacket.Name = n.Name
			dataPacket.Server = n.Server
			dataPacket.Port = n.Port
			dataPacket.LastLogIndex, dataPacket.LastLogTerm = n.lastLogIndexAndTerm()

			var byteSack bytes.Buffer
			enc := gob.NewEncoder(&byteSack)
			_ = enc.Encode(dataPacket)

			//setting mychannel to a fresh empty channel. Channel will recieve input if majority votes received
			mychannel = make(chan string, 1)

			//sends out vote request to all
			for i := 0; i < len(n.Myconnections); i++ {
				conn := n.ConnnectToNode(n.Myconnections[i])
				if conn != nil {
					conn.Write(byteSack.Bytes())
				}
			}

			//setting timeout to random number between 2*HeartBeatInterval to 3*HeartbeatInterval
			randomDuration := time.Duration(rand.Intn(HeartbeatInterval) + 2*HeartbeatInterval)

			select {
			//case majority votes are received
			case <-mychannel:
				//node becomes Leader
				n.MyRole = Leader
				//node resets vote map
				n.Votes = make(map[string]bool)
				//printing success story
				fmt.Println("I am Leader " + n.Name + " for Term " + strconv.Itoa(n.Term))
				break
			//case timeout occurs before a connection is received
			case <-time.After(randomDuration * time.Second):
				fmt.Println("Voting Time out As Candidate")
				break
			}

		case Leader:
			fmt.Println("I am Leader " + n.Name + " Now")
			n.SendHeartBeat() //Sends Heartbeat and AppendEntriesRPC
			time.Sleep(time.Duration(HeartbeatInterval) * time.Second)
		}
	}

}

// Wait : waits for further contact
func (n *Node) Wait() {
	ln, err := net.Listen("tcp", ":"+n.Port)
	if err != nil {
		log.Fatal(err)
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}

		go n.handleConnection(conn)
	}
}

func (n *Node) handleConnection(c net.Conn) {
	// buf := make([]byte, 4096)go run
	var dataPacket Protocol // buffer that holds converted bytes

	dec := gob.NewDecoder(c)       // a decoder that decodes dataPacket in connection
	err := dec.Decode(&dataPacket) // decoding dataPacket in connection
	if err != nil {
		log.Println(err)
		return
	}
	c.Close() //closing connection

	//Receiving Command From Client
	if dataPacket.Request[0] == "COMMAND" {
		c = n.ConnnectToNode(dataPacket.Port)
		fmt.Println("Command Received")
		if n.MyRole != Leader {
			c.Write([]byte("Contact Leader On This Port -> " + n.LeaderPort))
			return
		}
		//Append Log To Log File
		file, err := os.OpenFile("../logs/log"+n.Name+".txt", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		defer file.Close()
		if _, err := file.WriteString(strconv.Itoa(n.Term) + " " + dataPacket.Operation + " " + strconv.Itoa(dataPacket.Value) + "\n"); err != nil {
			log.Fatal(err)
		}
		c.Write([]byte("Log Updated, Wait For Sync & Response"))
		n.Log = append(n.Log, LogEntry{Term: n.Term, Operation: dataPacket.Operation, Value: dataPacket.Value})
		return
	}
	if dataPacket.Term < n.Term {
		return
	}
	if dataPacket.Request[0] == "PING" {
		mychannel <- "Response Recieved"

		fmt.Println("Ping from " + dataPacket.Name + " [" + dataPacket.Server + ":" + dataPacket.Port + "]")
		//case where leader of new term pings or leader port is not set
		if dataPacket.Term > n.Term || n.LeaderPort == "xxxx" || n.LeaderPort == "pending" {
			//term, leader port, and role reset
			n.Term = dataPacket.Term
			n.LeaderPort = dataPacket.Port
			n.MyRole = Follower
		} else if dataPacket.Term == n.Term && dataPacket.Port != n.LeaderPort {
			//case where an invalid leader pings. Unidentified leader pings for same term
			fmt.Println("Not responding because PING is not from True Leader")
			return
		}
		// Response To PING regarding the Log Entries [False By Default will become true if consistency check is valid]
		dataPacket.Request = []string{"PONG", "FALSE"}

		if dataPacket.PrevLogIndex == -1 ||
			(dataPacket.PrevLogIndex < len(n.Log) && dataPacket.PrevLogTerm == n.Log[dataPacket.PrevLogIndex].Term) {
			dataPacket.Request = []string{"PONG", "TRUE"}

			logInsertIndex := dataPacket.PrevLogIndex + 1
			newEntriesIndex := 0

			for {
				if logInsertIndex >= len(n.Log) || newEntriesIndex >= len(dataPacket.Entries) {
					break
				}
				if n.Log[logInsertIndex].Term != dataPacket.Entries[newEntriesIndex].Term {
					break
				}
				logInsertIndex++
				newEntriesIndex++
			}
			// At the end of this loop:
			// - logInsertIndex points at the end of the log, or an index where the
			//   term mismatches with an entry from the leader
			// - newEntriesIndex points at the end of Entries, or an index where the
			//   term mismatches with the corresponding log entry
			if newEntriesIndex < len(dataPacket.Entries) {
				fmt.Println("... inserting entries %v from index %d", dataPacket.Entries[newEntriesIndex:], logInsertIndex)
				n.Log = append(n.Log[:logInsertIndex], dataPacket.Entries[newEntriesIndex:]...)
				fmt.Println("... log is now: %v", n.Log)
			}
			// Set commit index.
			if dataPacket.LeaderCommit > n.CommitIndex {
				n.CommitIndex = Min(dataPacket.LeaderCommit, len(n.Log)-1)
				fmt.Println("... setting commitIndex = %d", n.CommitIndex)
			}
		}

		//constructing pong message
		pongPort := dataPacket.Port
		pongName := dataPacket.Name
		dataPacket.Name = n.Name
		dataPacket.Server = n.Server
		dataPacket.Port = n.Port
		dataPacket.Term = n.Term
		//sending pong message
		var byteSack bytes.Buffer
		enc := gob.NewEncoder(&byteSack)
		_ = enc.Encode(dataPacket)
		c = n.ConnnectToNode(pongPort)
		c.Write(byteSack.Bytes())
		c.Close()
		fmt.Println("Pinged back to " + pongName + " [True Leader for Term " + strconv.Itoa(n.Term) + "]")
		return
	} else if dataPacket.Request[0] == "PONG" { //leader receives this
		fmt.Println("Reply from " + dataPacket.Name + " [I AM ALIVE]")
		//If logs were inconsistent
		if n.MyRole == Leader && n.Term == dataPacket.Term {
			if dataPacket.Request[1] == "TRUE" {
				// Update NextIndex && MatchIndex here

			}
		}
	} else if dataPacket.Request[0] == "VOTE" { //sent by Candidate Node

		if dataPacket.Term > n.Term { //not becoming follower, why ?
			//updating term
			n.Term = dataPacket.Term
			//constructing vote message
			n.LeaderPort = "pending"
			replyPort := dataPacket.Port
			dataPacket.Request = []string{"VOTED"}
			dataPacket.Name = n.Name
			dataPacket.Server = n.Server
			dataPacket.Port = n.Port
			dataPacket.Term = n.Term
			//sending vote message
			var byteSack bytes.Buffer
			enc := gob.NewEncoder(&byteSack)
			_ = enc.Encode(dataPacket)
			c = n.ConnnectToNode(replyPort)
			c.Write(byteSack.Bytes())
			c.Close()
			fmt.Println("Vote Casted")
		}
	} else if dataPacket.Request[0] == "VOTED" && n.MyRole == Candidate { //received by Candidate
		fmt.Println("Vote Received")
		//setting vote in vote map for candidate
		n.Votes[dataPacket.Port] = true
		//sending success message to channel if majority votes received
		if len(n.Votes) > (len(n.Myconnections)+1)/2 {
			//Writing Current Term To LogFile
			file, err := os.OpenFile("../logs/log"+n.Name+".txt", os.O_RDWR, 0644)
			if err != nil {
				fmt.Println("failed opening file: %s", err)
			}
			defer file.Close()
			data := []byte(strconv.Itoa(n.Term) + "\n")
			_, err = file.WriteAt(data, 0) // Write at 0 beginning
			if err != nil {
				fmt.Println("failed writing Term to file: %s", err)
			}
			n.IntializeParams() //Initialize NextIndex[] & MatchIndex[]
			mychannel <- "You are Leader now"
		}
	}

}

func (n *Node) IntializeParams() {
	//Initializing NextIndex & MatchIndex [Used Only By Leader]
	n.NextIndex = make(map[string]int)
	n.MatchIndex = make(map[string]int)
	for i := 0; i < len(n.Myconnections); i++ {
		n.NextIndex[n.Myconnections[i]] = len(n.Log)
		n.MatchIndex[n.Myconnections[i]] = -1
	}
}

func (n *Node) SendHeartBeat() {

	////send out ping to all
	//for i := 0; i < len(n.Myconnections); i++ {
	//	conn := n.ConnnectToNode(n.Myconnections[i])
	//	if conn != nil {
	//		conn.Write(byteSack.Bytes())
	//	}
	//}
	//send out log Entries & Heartbeat to all
	for i := 0; i < len(n.Myconnections); i++ {
		go func(peerId string) {
			//Write Here
			temp := n.NextIndex[peerId]
			prevLogIndex := temp - 1
			prevLogTerm := -1
			if prevLogIndex >= 0 {
				prevLogTerm = n.Log[prevLogIndex].Term
			}
			var entries []LogEntry
			if len(n.Log)-1 >= n.NextIndex[peerId] {
				entries = n.Log[temp:]
			}

			//constructs new ping message to broadcast
			var dataPacket Protocol // buffer that holds converted bytes
			dataPacket.Request = []string{"PING"}
			dataPacket.Term = n.Term
			dataPacket.Name = n.Name
			dataPacket.Server = n.Server
			dataPacket.Port = n.Port
			dataPacket.PrevLogIndex = prevLogIndex
			dataPacket.PrevLogTerm = prevLogTerm
			dataPacket.Entries = entries
			dataPacket.LeaderCommit = n.CommitIndex

			c := n.ConnnectToNode(peerId)
			if c == nil {
				fmt.Println("error connecting to client sending AppendEntriesRPC")
				return
			}
			var byteSack bytes.Buffer
			enc := gob.NewEncoder(&byteSack)
			_ = enc.Encode(dataPacket)
			fmt.Println("Sending Ping to " + peerId)

		}(n.Myconnections[i])

	}
}

// lastLogIndexAndTerm returns the last log index and the last log entry's term
// (or -1 if there's no log)
func (n *Node) lastLogIndexAndTerm() (int, int) {
	if len(n.Log) > 0 {
		lastIndex := len(n.Log) - 1
		return lastIndex, n.Log[lastIndex].Term
	} else {
		return -1, -1
	}
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getBoolResponse() bool {
	var userInput string
	for userInput != "Y" && userInput != "N" {
		fmt.Scanln(&userInput)
		if userInput != "Y" && userInput != "N" {
			fmt.Println("Incorrect Input!! Reply with 'Y' or 'N':")
		}
	}
	return userInput == "Y"
}
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
