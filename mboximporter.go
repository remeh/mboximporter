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
    "time"

    "./src"
    "gopkg.in/mgo.v2"
    "github.com/bthomson/mbox"
)

func main() {
    var process mboximporter.Process

    // Flags 
    config := prepareFlags()

    // Opens a Mongo connection and prepare some tools to write
    mongo := mboximporter.GetConnection(config)
    defer mongo.Close()
    dao := mboximporter.NewMailDAO(config, mongo)

    // Prepare the channel we'll use as a queue of message to process
    messagesToDo := make(chan mail.Message, config.Concurrency)

    // Reads the data.
    messages, err := mbox.ReadFile(config.Filename, true)

    if err != nil {
        log.Println("Unable to open the file : ")
        log.Println(err)
        return
    }

    // Creates the workers
    var wg sync.WaitGroup
    var sem sync.WaitGroup
    for i := 0; i < config.Workers; i++ {
        wg.Add(1)

        go func(sem *sync.WaitGroup, dao *mboximporter.MailDAO) {
            // Each workers has its bulk to avoid
            // race condition between them.
            bulk := dao.NewBulk()
            bulk.Unordered()

            processed := 0 // number of processed messages by this worker
            for message := range messagesToDo {
                importMessage(config, dao, bulk, sem, &process, &message)
                processed++

                if processed % 500 == 0 {
                    // Executes the bulk
                    bulk.Run()
                    // Creates a new one for this worker
                    bulk = dao.NewBulk()
                    bulk.Unordered()
                    log.Prinltn("Bulk wrote.")
                }

                // Last inserts
                bulk.Run()
            }
            wg.Done()
        }(&sem, dao)
    }

    // Amount to import
    log.Printf("%d messages to import.\n", len(messages))
    maxMessages := len(messages)
    countToImport := config.Count
    if config.Count == -1 {
        countToImport = maxMessages
    }
    for i := 0; i < countToImport; i++ {
        messagesToDo <- *messages[i] // Enqueue the message to be processed
        sem.Add(1)
    }

    log.Println("Working.")
    sem.Wait()
    log.Printf("Processed %d messages :", process.ProcessedMessages+process.IgnoredChatMessages)
    log.Printf("- Imported %d messages.", process.ProcessedMessages)
    log.Printf("- Ignored %d chat messages.", process.IgnoredChatMessages)
    log.Printf("- Errored on %d messages.", countToImport - (process.ProcessedMessages+process.IgnoredChatMessages))
    log.Println("End of execution.")
}

// Prepares the CLI flags for the
// Mongo connection and the file to import.
func prepareFlags() mboximporter.Config {
    mongoURI := flag.String("m", "localhost", "The Mongo URI to connect to MongoDB.")
    dbName := flag.String("d", "mails", "The DB name to use in MongoDB.")
    filename := flag.String("f", "mails.mbox", "Name of the filename to import")
    workers := flag.Int("w", 10, "Maximum amount of workers.")
    concurrency := flag.Int("c", 20, "Maximum amount of messages in the same time in the pool of process.")
    count := flag.Int("n", -1, "Number of mails to import.")

    flag.Parse()

    return mboximporter.Config{MongoURI: *mongoURI, DBName: *dbName, Count: *count, Filename: *filename, Concurrency: *concurrency, Workers: *workers}
}

func importMessage(c mboximporter.Config, dao *mboximporter.MailDAO, bulk *mgo.Bulk, sem *sync.WaitGroup, process *mboximporter.Process, msg *mail.Message) {
    defer sem.Done()

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
                log.Println(err)
                return
            }
        } else if k == "To" {
            recipients = v
        } else if k == "Subject" {
            subject = v[0] // FIXME could have many values
            if err != nil {
                log.Println("Unable to unescape the subject.")
                log.Println(err)
                return
            }
        } else if k == "Content-Type" {
            contentType = v[0]
        } else if k == "X-Gmail-Labels" && v[0][0:4] == "Chat" {
            // Ignore chat messages from GMail
            process.IgnoredChatMessages++
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
        log.Println("Unable to read the type of the content.")
        log.Println(err)
        return
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
                log.Println(err)
                return
            }
            slurp, err := ioutil.ReadAll(p)
            if err != nil {
                log.Println(err)
                return
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
        log.Println(err)
        return
    }

    importMsg := &mboximporter.Mail{
        Headers: headers,
        Sender: sender,
        Recipients: recipients,
        Date: date,
        Subject: subject,
        Body: finalBody }

    // Saves the writes into the Bulk
    bulk.Insert(importMsg)

    time.Sleep(15)

    process.ProcessedMessages++
    //log.Println("\"" + importMsg.Subject + "\" imported.")
}
