We have used veroius websites and ChatGPT to help us problemsolve, when TA's weren't avaiable

For being able to even start the program, you need to open the Handin4 folder in visual studio code :D

For 3 Peers:
- make 3 bash terminals
- all of them needs to find its way into the peer folder using "cd <the folder>" and "cd .." to move forward and backwords
- in the 3 peer terminals, write: "go run peer.go -name <1, 2, or 3 respectivly> -port <5001, 5002, 5003 respectivly>"

when the peer are connected, they can add other ports (service discovery) and join the tokenring:
	Use /add <port> to join the chatroom
		all peers need to add the other two ports 
		( example on peer with port 5001 --> 
			/add 5002
			/add 5003 )
	Use /start in all terminals, respecticly with starting port 5001 last

The server keeps a log of alle interactiong called log.txt, and it is under the folder peer

