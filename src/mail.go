package mboximporter

import (
    "time"

    "labix.org/v2/mgo/bson"
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
