package messages

// https://gobyexample.com/channel-directions
type OutgoingQueue chan<- Message
type IncomingQueue <-chan Message
