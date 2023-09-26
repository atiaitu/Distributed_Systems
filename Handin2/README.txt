a) What are packages in your implementation? What data structure do you use to transmit data and meta-data?
- In the program the channels are used to send to and from a client to a server.

b) Does your implementation use threads or processes? Why is it not realistic to use threads?
- The program uses go routines which starts concurrent and communicates via channels.
- This is done to replicate the 3-way handshake for TCP


c) In case the network changes the order in which messages are delivered, how would you handle message re-ordering?
- This case is not implemented in the program

d) In case messages can be delayed or lost, how does your implementation handle message loss?
- In this case our implementation handles this problem via a time.
- For each appending message, if it hasn’t been delivered in a given timespan, it simply sends a message given the information that a timeout has been reached
- In the case where the message doesn’t match, it sendt a message given the information that a wrong code has been send

e) Why is the 3-way handshake important?
- The reason why a 3-way handshake is important in a TCP is that it ensures that both the sender and receiver are ready and willing to communicate 
- The handshake helps establish a reliable connection by confirming that both parties are actively participating and are aware of each other's state
- It can provide a level of security by ensuring that both devices are legitimate and authorized to communicate
