# HerdiusEngineeringTask

## Requirements
* Clients must send stream of integer messages at any time to server.
* Server will respond over another stream every time the client has sent a new maximum integer, with that new max.
* Stream to server should be encrypted and its origin verifiable.

## Solution
* Use GRPC with Bi-directional streaming mode.
* We can use TLS in GRPC to encrypt the data sent to the server.
* We can use mutual TLS mode in GRPC to authenticate the client as well.
* We can then use the client's public key to identify the request.
