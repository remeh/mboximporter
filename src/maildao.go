package mboximporter

import (
    "encoding/json"
    "fmt"
    "log"

    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"

    "github.com/lib/pq"
    "database/sql"
)

const (
    COLLECTION_MAIL = "mail"
)

// ---------------------- 
// Declarations

type MailDAO interface {
    Save(mail *Mail) error
}

// DAO for Mail collection in Mongo
type MailMongoDAO struct {
    mongo       *Mongo
    collection  *mgo.Collection
}

// DAO for Mail table in PGSQL
type MailPGDAO struct {
    DB *sql.DB
    PreparedInsert *sql.Stmt
}

// ---------------------- 
// Methods

// Mongo

func NewMailMongoDAO(c Config, m *Mongo) *MailMongoDAO {
    return &MailMongoDAO{m, m.GetCollection(c, COLLECTION_MAIL)}
}

func (d *MailMongoDAO) Save(mail *Mail) error {
    if (len(mail.Id) > 0) {
        return d.collection.Update(bson.M{"_id": mail.Id}, mail)
    } else {
        return d.collection.Insert(mail)
    }
}

// PostgreSQL

func NewMailPGDAO(c Config) *MailPGDAO {
    // Builds the connection string
    connectionString := fmt.Sprintf("user=%s dbname=%s sslmode=disable", c.PostgresUser, c.PostgresDatabase)
    if len(c.PostgresPassword) > 0 {
        connectionString += " password="+c.PostgresPassword
    }

    // Open the actual connection
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        log.Println("Unable to open the PostgreSQL connection.")
        log.Fatal(err)
    }

    // Prepare a query
    stmt, err := db.Prepare(fmt.Sprintf("INSERT INTO %s.%s VALUES ($1)", pq.QuoteIdentifier(c.PostgresSchema), pq.QuoteIdentifier(c.PostgresTable)))

    if err != nil {
        log.Fatal(err)
    }

    log.Println("Opened a PG connection.")

    return &MailPGDAO{DB: db, PreparedInsert: stmt}
}

func (d *MailPGDAO) Save(mail *Mail) error {
    json, err := json.Marshal(*mail)
    if err != nil {
        log.Fatal(err)
    }
    _, err = d.PreparedInsert.Exec(json)
    if err != nil {
        log.Fatal(err)
    }
    return err

}
