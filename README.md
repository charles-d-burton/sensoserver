## Synopsis

This is a server program to injest sensor data and publish it to Firebase Cloud Messaging from Google.  You register tokens(Android/IOS), this returns a topic that you can then send data to from your sensors.  The data is schemaless so you can send whatever you like.

## Motivation

I conceived of this because I wanted a way to exert command and monitoring over various sensors throughout my house.  For right now it's not possible to send data back to a sensor but I'm working on that.

## Installation
go get the following packages:

"github.com/satori/go.uuid"
"github.com/NaySoftware/go-fcm"
"rsc.io/letsencrypt"
"github.com/urfave/cli"

Simply clone the repo and run with go run main.go --ssl off --port 8081

## License
This software is licensed under the MIT license.
