package mboximporter

import (
    "time"

    "gopkg.in/mgo.v2/bson"
)

type Mail struct {
    Id bson.ObjectId `bson:"_id,omitempty"`

    Headers []string
    Sender string
    Recipients []string
    Date time.Time
    Subject string
    Body string
}
