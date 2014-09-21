# MboxImport

A basic tool to import your `.mbox` file into MongoDB for query.

It's just a basic side-project development developed in a few hours for fun. Don't hesitate to provide pull-requests or to submit ideas.

## Build

```
go get github.com/bthomson/mbox
go get labix.org/v2/mgo
go build mailimporter.go

```

## Usage

```
Usage of ./mailimporter:
    -c=-1: Number of mails to import.
    -d="mails": The DB name to use in MongoDB.
    -f="mails.mbox": Name of the filename to import
    -m="localhost": The Mongo URI to connect to MongoDB.
```
