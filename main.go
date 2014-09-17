package main

import (
    "fmt"
    "io/ioutil"
    "net/mail"
//  "net/url"

    "./mailimport"
    "github.com/bthomson/mbox"
    "github.com/remeh/mail/quotedprintable"
)


func main() {
    // Reads the data.
    messages, err := mbox.ReadFile("./mails.mbox", true)

    if err != nil {
        fmt.Println("Unable to open the file : ")
        fmt.Println(err)
        return
    }

    // Opens a Mongo connection
    mongo := mailimport.GetConnection()
    defer mongo.Close()


    // Import all the messages.
    if len(messages) != 0 {
        for i := 0; i < 1000; i++ {
            go importMessage(mongo, messages[i])
        }
    }

    c := make(chan int)
    <-c
}

// FIXME error handling
func decodeString(content string) string {
    /*
    ud, err := url.QueryUnescape(content)
    if err != nil {
        fmt.Println(err)
        return ""
    }
    */

    dec, err := quotedprintable.DecodeString(content)
    if err != nil {
        fmt.Println(err)
        return content
    }
    return string(dec)
}

func importMessage(mongo *mailimport.Mongo, msg *mail.Message) {
    // Export headers
    headers := make([]string,len(msg.Header))
    var err error
    var sender string
    var subject string

    i := 0
    for k, v := range msg.Header {
        // Specific header
        if k == "From" {
            sender = decodeString(v[0]) // FIXME could have many values
            if err != nil {
                fmt.Println("Unable to unescape the sender.")
                fmt.Println(err)
                continue
            }
        }
        if k == "Subject" {
            subject = decodeString(v[0]) // FIXME could have many values
            if err != nil {
                fmt.Println("Unable to unescape the subject.")
                fmt.Println(err)
                continue
            }
        }

        // Ignore chat messages.
        if k == "X-Gmail-Labels" && v[0][0:4] == "Chat" {
            return
        }

        // Others
        stringValue := k +": "+msg.Header.Get(k)
        headers[i] = stringValue
        i++
    }

    // Recipients
    // TODO
    recipients := make([]string, 0)

    // Reads the body
    body, err := ioutil.ReadAll(msg.Body)
    if err != nil {
        fmt.Println("Unable to read a body.")
        fmt.Println(err)
        return
    }
    finalBody := decodeString(string(body))

    // Reads the date
    date, err := msg.Header.Date()
    if err != nil {
        fmt.Println("Unable to read the date.")
        fmt.Println(err)
        return
    }

    importMsg := &mailimport.Mail{
        Headers: headers,
        Sender: sender,
        Recipients: recipients,
        Date: date,
        Subject: subject,
        Body: finalBody }

    //fmt.Printf("Headers: %s\n", importMsg.Headers)
    fmt.Printf("Date: %s\n", importMsg.Date)
    fmt.Println("Sender: " + importMsg.Sender)
    fmt.Println("Title: " + importMsg.Subject)
    fmt.Println("")

    // Saves in MongoDB
    dao := mailimport.NewMailDAO(mongo)
    dao.Save(importMsg)
}
