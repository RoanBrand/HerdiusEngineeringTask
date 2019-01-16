# HerdiusEngineeringTask

## Requirements
* Clients must send stream of integer messages at any time to server
* Server will respond over another stream every time the client has sent a new maximum integer, with that new max
* Stream to server should be encrypted and its origin verifiable.

## Solution
* Use GRPC with B-directional streaming mode.
* Use Public/Private key encryption to encrypt and sign the data.
* Server should have a list of known clients and their public keys.
