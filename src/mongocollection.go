package mboximporter

import (
    "gopkg.in/mgo.v2"
)

// ---------------------- 
// Declarations

// A connection must be able to return a Mongo collection.
// @author Rémy MATHIEU
type MongoConnection interface {
    GetCollection(string) *mgo.Collection
}
