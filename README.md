# RAFT
This is an implementation of raft in Golang

RAFT Documentation
# How Do You Run Our Project ?
You’ll have to open at least 2 Terminals in the project's main directory.  
Run The main.go file with parameters in following order  
Server Name [Should Only Be Capital Alphabets i.e ‘A’, ‘B’]  
Capital Alphabet names are necessary to be consistent with our
log Files, because our log files are named logA, logB …
For now we have created 5 log files if you want to test the project
With more than 5 nodes you can add the log files in logs directory.
Listening Port   
Dialing Ports [We do not read the topology information from any file,
 you make the topology on genesis through dialing ports and listening ports].  
An example with 3 nodes is given below  
Node A: go run main.go A 2000 2001 2002   
Node B: go run main.go B 2001 2000 2002   
Node C: go run main.go C 2002 2000 2001   

Run the client.go file with parameters in following order
Client Name [Can’t contain spaces]
Listening Port
Example: go run client.go Bill 3000
After the client is up and running you can connect to the leader and send in commands which would then be stored in the leader's log  and replicated throughout the cluster. If you connect to a server who is not a leader he will respond back with the leader’s connection details and would ask the client to connect to the leader instead. The server could just transfer the command to the leader but this is a design choice which does not affect the working of the RAFT algorithm itself.

*GET command is never logged.


# Partitioning, Recovery & Test Cases:
To simulate partitioning we just kill the desired server(s) and the rest of the cluster goes on. Killing the server is safe because our logs are persistent and are stored in a persistent storage (Log Files). When a server restarts it reads the term and log entries in it’s log file and is initialized to the state it was killed on.
To recover a server just start that server again using the same name, listening port and dialing ports it had on genesis.
 The log files are being updated on run time for all the servers.

# Test Cases:
We have tested our code on all the cases we could think of and it’s working as desired. Some of the cases are described below  
Case-1: A Follower Crashes And The Leader Remains Same
If a follower crashes it will miss out on all the commands being committed and sent all over the cluster, however before it’s crash it stored all of it’s log entries and it’s term to a persistent storage, upon restarting it will initialize with the stored  term and log from the log file. It will then receive the AppendRPC which would have the PrevLogIndex equal to the preceding latest log entry and the server would insert all the incoming entries to it’s log.
  
Case-2: A Follower Crashes And The Leader Changes
If the leader changes on restart the leader would initialize the NextIndex for this follower based on it’s last log entry and update the PrevLogTerm accordingly, it will then send the AppendRPC to the follower and force the follower to replicate the log of the leader because of the consistency check no matter if the follower had entries from the previous leader which it didn’t get the chance to respond back on.
  
Case-3: A Leader Crashes
If a leader crashes it will restart with the last term and log entries it had before crash and would still send AppendRPC until it receives a heartbeat with a term greater than his to  which then he will become a follower and update it’s log according to the current leader’s log, the previous leader’s AppendRPC’s would just be rejected by the servers. Our previous servers won’t have extra bogus entries because we are not submitting commands in batches, instead we submit and replicate only a single command at a time. So our algorithm is safe from redundant entries in a previous leader’s log which were not yet sent to the followers then.
  
 


