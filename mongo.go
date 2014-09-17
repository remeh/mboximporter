package mailimport

import (
    "labix.org/v2/mgo"
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
// TODO pool of a sessions / connections
func GetConnection() *Mongo {
    m := new(Mongo)
    session, err := mgo.Dial("localhost, localhost")
    if err != nil {
        panic(err)
    }
    m.session = session
    return m
}

func (m *Mongo) GetCollection(name string) *mgo.Collection {
    // TODO db name
    return m.session.DB("mails").C(name)
}

func (m *Mongo) Close() {
    m.session.Close()
}
