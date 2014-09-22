package mboximporter

import (
    "gopkg.in/mgo.v2"
)

// ---------------------- 
// Declarations

// A Mongo connection
// @author Rémy MATHIEU
type Mongo struct {
    session *mgo.Session
    database *mgo.Database
}

// ---------------------- 
// Methods

// Retrieves a new Mongo Connection.
func GetConnection(c Config) *Mongo {
    m := new(Mongo)
    session, err := mgo.Dial(c.MongoURI)
    if err != nil {
        panic(err)
    }
    m.session = session
    return m
}

func (m *Mongo) GetCollection(c Config, name string) *mgo.Collection {
    return m.session.DB(c.DBName).C(name)
}

func (m *Mongo) Close() {
    m.session.Close()
}
