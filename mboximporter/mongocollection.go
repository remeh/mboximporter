package mboximporter

import (
    "labix.org/v2/mgo"
)

// ---------------------- 
// Declarations

// A connection must be able to return a Mongo collection.
// @author Rémy MATHIEU
type MongoConnection interface {
    GetCollection(string) *mgo.Collection
}
