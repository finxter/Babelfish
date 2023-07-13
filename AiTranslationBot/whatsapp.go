package main

import (
    //"bufio"
    "context"
    //"encoding/hex"
    //"encoding/json"
    "errors"
    "flag"
    "fmt"
    //"mime"
    //"net/http"
    "os"
    //"os/signal"
    //"strconv"
    "strings"
    "sync/atomic"
    //"syscall"
    "time"

    _ "github.com/mattn/go-sqlite3"
    "github.com/mdp/qrterminal/v3"
    "google.golang.org/protobuf/proto"

    "go.mau.fi/whatsmeow"
    //"go.mau.fi/whatsmeow/appstate"
    //waBinary "go.mau.fi/whatsmeow/binary"
    waProto "go.mau.fi/whatsmeow/binary/proto"
   // "go.mau.fi/whatsmeow/store"
    "go.mau.fi/whatsmeow/store/sqlstore"
    "go.mau.fi/whatsmeow/types"
    "go.mau.fi/whatsmeow/types/events"
    waLog "go.mau.fi/whatsmeow/util/log"
)



var cli *whatsmeow.Client
var Log waLog.Logger

var logLevel = "INFO"
var dbDialect = flag.String("db-dialect", "sqlite3", "Database dialect (sqlite3 or postgres)")
var dbAddress = flag.String("db-address", "file:mdtest.db?_foreign_keys=on", "Database address")
var pairRejectChan = make(chan bool, 1)


func whatsAppStart(){

    flag.Parse()
    Log = waLog.Stdout("Main", logLevel, true)
    dbLog := waLog.Stdout("Database", logLevel, true)

    // Define a new db and try to read if there is any device
    storeContainer, err := sqlstore.New(*dbDialect, *dbAddress, dbLog)
    if err != nil {
        Log.Errorf("Failed to connect to database: %v", err)
        return
    }
    device, err := storeContainer.GetFirstDevice()
    if err != nil {
        Log.Errorf("Failed to get device: %v", err)
        return
    }


    cli = whatsmeow.NewClient(device, waLog.Stdout("Client", logLevel, true))

    // This variable is used to indicate whether the system is currently waiting for a pair.

    var isWaitingForPair atomic.Bool

    // Overall, this code represents a callback function that handles pre-pairing logic.
    // It sets a flag to indicate that the system is waiting for a pair, logs pairing
    // details, and waits for a rejection signal or a timeout. Depending on the received
    // signal, it decides whether to accept or reject the pair.

    cli.PrePairCallback = func(jid types.JID, platform, businessName string) bool {
        // indicates that the system is waiting for a pair.
        isWaitingForPair.Store(true)
        // to ensure that isWaitingForPair is set back to false when the function exits, regardless of the return path.
        defer isWaitingForPair.Store(false)

        Log.Infof("Pairing %s (platform: %q, business name: %q). Type r within 3 seconds to reject pair", jid, platform, businessName)
        select {

            case reject := <-pairRejectChan:
                if reject {
                    Log.Infof("Rejecting pair")
                    return false
                }
            case <-time.After(3 * time.Second):
        }
        Log.Infof("Accepting pair")
        return true
    }

    // below code is responsible for obtaining a QR channel, handling the events 
    // received from the channel, and performing appropriate actions based on the 
    // event type. Basically generates QR code for pairing this bot/app with your mobile device

    ch, err := cli.GetQRChannel(context.Background())
    if err != nil {
        // This error means that we're already Log.ed in, so ignore it.
        if !errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
            Log.Errorf("Failed to get QR channel: %v", err)
        }
    } else {
        go func() {
            for evt := range ch {
                if evt.Event == "code" {
                    qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
                } else {
                    Log.Infof("QR channel result: %s", evt.Event)
                }
            }
        }()
    }

    cli.AddEventHandler(handler)
    err = cli.Connect()
    if err != nil {
        Log.Errorf("Failed to connect: %v", err)
        return
    }

}


func handler(rawEvt interface{}) {
    switch evt := rawEvt.(type){
        case *events.Message:{
            metaParts := []string{fmt.Sprintf("pushname: %s", evt.Info.PushName), fmt.Sprintf("timestamp: %s", evt.Info.Timestamp)}
            Log.Infof("Received message %s from %s (%s): %+v", evt.Info.ID, evt.Info.SourceString(), strings.Join(metaParts, ", "), evt.Message)
            // pass received msg to the frontend
            // User = source phone number
            ReceiveMsg(evt.Info.Sender.User, evt.Message.GetConversation())
        }
    }
}

// JID = Jabber Identifier. Its part of whatsapp messaging system and its format
// phone_number@s.whatsapp.net which is the default format

func parseJID(arg string) (types.JID, bool) {

    if arg[0] == '+' {
        arg = arg[1:]
    }
    if !strings.ContainsRune(arg, '@') {
        return types.NewJID(arg, types.DefaultUserServer), true
    } else {
        recipient, err := types.ParseJID(arg)
        if err != nil {
            Log.Errorf("Invalid JID %s: %v", arg, err)
            return recipient, false
        } else if recipient.User == "" {
            Log.Errorf("Invalid JID %s: no server specified", arg)
            return recipient, false
        }
        return recipient, true
    }
}


// handleCmd will actually handle the sending of the message to the number 
// using whatsapp protocol implemented in the framework - whatsmeow

// We can as well extend the handleCmd to handle more complex scenarios such as 
// 1. sending messages to a group of people i.e multisend (like you have a whatsapp group)
// 2. Get all users
// 3. Get all groups and subgroups
// 4. send images 
// 5. archive messages 
// etc and so on
func handleCmd(cmd string, args []string) {

    switch cmd {

        case "send":
            if len(args) < 2 {
                Log.Errorf("Usage: send <jid> <text>")
                return
            }
            recipient, ok := parseJID(args[0])
            if !ok {
                return
            }
            msg := &waProto.Message{Conversation: proto.String(strings.Join(args[1:], " "))}
            resp, err := cli.SendMessage(context.Background(), recipient, msg)
            if err != nil {
                Log.Errorf("Error sending message: %v", err)
            } else {
                Log.Infof("Message sent (server timestamp: %s)", resp.Timestamp)
            }
    }
}









/*
        case "reconnect":
            cli.Disconnect()
            err := cli.Connect()
            if err != nil {
                Log.Errorf("Failed to connect: %v", err)
            }

    c := make(chan os.Signal)
    input := make(chan string)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        defer close(input)
        scan := bufio.NewScanner(os.Stdin)
        for scan.Scan() {
            line := strings.TrimSpace(scan.Text())
            if len(line) > 0 {
                input <- line
            }
        }
    }()
    for {
        select {
        case <-c:
            Log.Infof("Interrupt received, exiting")
            cli.Disconnect()
            return
        case cmd := <-input:
            if len(cmd) == 0 {
                Log.Infof("Stdin closed, exiting")
                cli.Disconnect()
                return
            }
            if isWaitingForPair.Load() {
                if cmd == "r" {
                    pairRejectChan <- true
                } else if cmd == "a" {
                    pairRejectChan <- false
                }
                continue
            }
            args := strings.Fields(cmd)
            cmd = args[0]
            args = args[1:]
            go handleCmd(strings.ToLower(cmd), args)
        }
    }
    */


