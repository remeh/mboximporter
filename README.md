# MboxImporter

A basic tool to import your `.mbox` file into MongoDB for query.

It's just a basic side-project development developed in a few hours for fun. Don't hesitate to provide pull-requests or to submit ideas.

On my machine (core i7, 8gb of ram), it imports approximately `2500 mails/second`.

## Build

```
go get github.com/bthomson/mbox
go get labix.org/v2/mgo
go build mboximporter.go

```

## Usage

```
Usage of ./mboximporter:
    -c=20: Maximum amount of messages in the same time in the pool of process.
    -d="mails": The DB name to use in MongoDB.
    -f="mails.mbox": Name of the filename to import
    -m="localhost": The Mongo URI to connect to MongoDB.
    -n=-1: Number of mails to import.
    -w=10: Maximum amount of workers.
```
