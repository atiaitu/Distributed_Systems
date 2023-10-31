For being able to even start the program, you need to open the Handin3 folder in visual studio code :D

For 1 server and 3 clients:
- make 4 bash terminals
- 3 of them needs to find its way into the client folder using "cd <something>" and "cd .." to move forward and backwords
- the last one needs to find its way into the server folder
- in the server terminal, write: "go run server.go -port 5400"
- in the 3 client terminals, write: "go run client.go -name <name> -server 5400"

when the clients are connected to the server, they can join, leave and publish messages using these commands:
	Use /j to join the chatroom
	Use /m <your message> to write a message
	Use /l to leave the chatroom

The server keeps a log of alle interactiong called log.txt, and it is under the folder server

