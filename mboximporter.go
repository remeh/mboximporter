package main

import (
    "flag"
    "io"
    "io/ioutil"
    "log"
    "mime"
    "mime/multipart"
    "net/mail"
    "strings"
    "sync"

    "./mboximporter"
    "github.com/bthomson/mbox"
)

func main() {
    // Flags 
    config := prepareFlags()

    // Reads the data.
    messages, err := mbox.ReadFile(config.Filename, true)

    if err != nil {
        log.Println("Unable to open the file : ")
        log.Println(err)
        return
    }

    // Opens a Mongo connection
    mongo := mboximporter.GetConnection(config)
    defer mongo.Close()

    // Some info on numbers
    log.Printf("%d messages to import.\n", len(messages))
    maxMessages := len(messages)
    countToImport := config.Count
    if config.Count == -1 {
        countToImport = maxMessages
    }

    // Our semaphore
    var sem sync.WaitGroup

    // Do the actual work of importing the mails.
    if len(messages) != 0 {
        for i := 0; i < countToImport; i++ {
            sem.Add(1)
            go importMessage(config, mongo, &sem, messages[i])
        }
    }

    log.Println("Working.")
    sem.Wait()
    log.Println("End of execution.")
}

// Prepares the CLI flags for the
// Mongo connection and the file to import.
func prepareFlags() mboximporter.Config {
    mongoURI := flag.String("m", "localhost", "The Mongo URI to connect to MongoDB.")
    dbName := flag.String("d", "mails", "The DB name to use in MongoDB.")
    filename := flag.String("f", "mails.mbox", "Name of the filename to import")
    count := flag.Int("c", -1, "Number of mails to import.")

    flag.Parse()

    return mboximporter.Config{MongoURI: *mongoURI, DBName: *dbName, Count: *count, Filename: *filename}
}

func importMessage(c mboximporter.Config, mongo *mboximporter.Mongo, sem *sync.WaitGroup, msg *mail.Message) {
    // Export headers
    headers := make([]string,len(msg.Header))
    var err error
    var sender string
    var subject string
    var recipients []string
    contentType := "plain/text"

    i := 0
    for k, v := range msg.Header {
        // Specific header
        if k == "From" {
            sender = v[0] // FIXME could have many values
            if err != nil {
                log.Println("Unable to unescape the sender.")
                log.Fatal(err)
            }
        }
        if k == "To" {
            recipients = v
        }
        if k == "Subject" {
            subject = v[0] // FIXME could have many values
            if err != nil {
                log.Println("Unable to unescape the subject.")
                log.Fatal(err)
            }
        }
        if k == "Content-Type" {
            contentType = v[0]
        }

        // Ignore chat messages from GMail
        if k == "X-Gmail-Labels" && v[0][0:4] == "Chat" {
            return
        }

        // Others
        stringValue := k +": "+msg.Header.Get(k)
        headers[i] = stringValue
        i++
    }

    // Creates a reader.
    mediaType, params, err := mime.ParseMediaType(contentType)
    if err != nil {
        log.Fatal("Unable to read the type of the content.")
    }
    reader := multipart.NewReader(msg.Body, params["boundary"])

    // Reads the body
    finalBody := ""
    if strings.HasPrefix(mediaType, "multipart/") {
        for {
            p, err := reader.NextPart()
            if err == io.EOF {
                break
            }
            if err != nil {
                log.Fatal(err)
            }
            slurp, err := ioutil.ReadAll(p)
            if err != nil {
                log.Fatal(err)
            }
            finalBody += string(slurp)
        }
    } else {
        txt, err := ioutil.ReadAll(msg.Body)
        if err != nil {
            log.Fatal(err)
        }
        finalBody += string(txt)
    }

    // Reads the date
    date, err := msg.Header.Date()
    if err != nil {
        log.Println("Unable to read the date.")
        log.Fatal(err)
        return
    }

    importMsg := &mboximporter.Mail{
        Headers: headers,
        Sender: sender,
        Recipients: recipients,
        Date: date,
        Subject: subject,
        Body: finalBody }

    // Saves in MongoDB
    dao := mboximporter.NewMailDAO(c, mongo)
    dao.Save(importMsg)

    log.Println("\"" + importMsg.Subject + "\" imported.")
    sem.Done()
}
