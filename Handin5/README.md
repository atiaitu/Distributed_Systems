We have used veroius websites and ChatGPT to help us problemsolve, when TA's weren't avaiable

For being able to even start the program, you need to open the Handin5 folder in visual studio code :D

For 3 server and 2 clients:
- make 5 bash terminals
- 3 of them needs to find its way into the server folder using "cd <something>" and "cd .." to move forward and backwords
- the other 2 needs to find its way into the client folder
- in the first server terminal, write: "go run server.go -port 5400 -name Server0 -wipe" //The -wipe just 	wipes the log
- in the second server terminal, write "go run server.go -port 5401 -name Server1"
- in the third server terminal, write "go run server.go -port 5402 -name Server2"
- in the first client terminal, write: "go run client.go -name <name>"
- in the second client terminal, write: "go run client.go -name <otherName>"

when the input above has been executed, clients can bit and review the stauts of the auction using these commands:
	Use /b <your bid> to place a bit
	Use /r to review the status of the auction

- The server keeps a log of all interactions called log.txt, and it is under the server folder
- Try to crash one server, and becuase of the construction of the servers, the auction will still function as expected.
- Right now, the auction is set to 30 seconds, and starts when the first bid is placed, if you want to change timer, see line 212 in the server.go code.


