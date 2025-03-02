# chat-server

This repository defines a chat server using TCP connections to communicate with users.

# Badges

[![Build services](https://github.com/Knoblauchpilze/chat-server/actions/workflows/build-and-push.yml/badge.svg)](https://github.com/Knoblauchpilze/chat-server/actions/workflows/build-and-push.yml)

[![codecov](https://codecov.io/gh/Knoblauchpilze/chat-server/graph/badge.svg?token=0ABFMO9WVY)](https://codecov.io/gh/Knoblauchpilze/chat-server)

# How does this work?

## Generalities

This project defines a server allowing users to connect to various rooms and chat with other registered users in private rooms (i.e. accessible only to two users). The chat server offers persistent storage of messages in the form of a chat history.

Chat rooms can be created by users and should have a unique name. A user is free to join a room or leave it.

## Technology overview

The chat server uses persistent TCP connections to receive messages from users and send them messages concerning them. The connection can either be closed by the user (which the server detects) or by the server in case the user misbehave.

Messages, rooms and concepts of the server are saved in a database to guarantee the persistence of the data.

## The TCP server

The general idea is that the server is listening on a specific port to which clients can connect to. As soon as a connection is established, some callbacks are triggered to know if the user should be denied or granted the connection.

The connection stays open for the duration of the session and is used as a bidirectional communication channel between the server and the user.

We want to handle graceful shutdown of the server. This means making sure that all connections are properly terminated (with a message sent to the client). Additionally this also means persisting any state to the database before closing the server. This [tutorial](https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/) is useful to describe the mechanism we followed.

## Connection lifecycle

When a connection has been accepted, the server listens to events happening on it. This typically includes:

- data received from the user
- disconnection

On top of this there are internal events that will produce data to send through the connection. This mainly includes messages for rooms that the user belongs to or direct messages from other users.

## Processing of messages

The diagram below presents the architecture of the server and how it handles messages.

![Server architecture](resources/server-architecture.png)

The clients are initially received by the `ConnectionAcceptor`, which forwards the connection request to the `ConnectionManager` which in turn forwards it to the `ClientManager`.

The `ClientManager` is the central authority keeping track of the clients connected to a node of the chat server. It also performs the filtering to allow/deny connection requests from certain clients (typically if a client is banned).

It also handles the handshake with new clients: this means trying to communicate in a predefined way to establish that the conneciton is legit and that we have a genuine client trying to interact with the chat and not some attacker/bot probing the connection.

When the connection is approved, the `ClientManager` will send a message to the internal message queue (see following paragraphs) and the `ConnectionManager` will start a dedicated process to receive data from the client.

Each time such data is received, the `MessageParser` will be notified and will try to decode it: if it fails to do so, the connection will be terminated. If it succeeds, the message will be sent to the `InputMessageQueue`.

The `InputMessageQueue`'s role is to dispatch the messages and to allow their processing. It's not doing any processing itself but rather acts as a synchronization mechanism bewteen the `ClientManager`/`MessageParser` and the `MessageProcessor`s.

The `MessageProcessor`s are designed to be executed concurrently: they each listen to the message queue and grab messages as they appear. They handle the necessary processing for the message which includes:

- persisting information to the database
- creating new messages
- route the messages to the concerned clients

Depending on the message some of the above operations might not happen.

When a message needs to be sent to the customer, the `MessageProcessor` forwards the call to the `ClientManager` as it is the only actor which knows about the connections.

At any point if a client disconnects, the `ConnectionManager` will notify the `ClientManager`, which will take the appropriate actions to reflect this in the database and notify other clients.

# Ideas

- Deactivate rooms if nobody is in them anymore
- Handle private/public rooms
- Do not allow users to leave private rooms (or delete them)
- Implement invitation to a room
- Login and logout system
