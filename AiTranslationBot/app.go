package main

import (
    "context"
    "fmt"
    "github.com/wailsapp/wails/v2/pkg/runtime"
    "strings"
)
// [0] = number
// [1] = actäual msg to send
var sendargs = make([]string, 2)
var translateTo string // store language to be translated

// These variables are needed for receiving the msgs
var rcvargs = make([]string, 1)
var runtimectx context.Context // need context for recieving msgs


// App struct
type App struct {
    ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
    return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
    runtimectx = ctx
}

// Initialize whatsapp bot
func (a *App) InitWhatsApp(language string) {
    fmt.Printf("Init whatsapp bot and language: %s\n", language)
    translateTo = language
    whatsAppStart()
}


func (a *App) SendNumber(num string) bool{
    fmt.Printf("This is the number to send/recv msgs %s!\n",num)
    sendargs[0] = num
    return true
}

func (a *App) SendMsg(msg string) bool{
    fmt.Printf("original msg to send: %s\n", msg)

    // translate to english before sending over whatsapp using OpenAI - chatGPT
    s := OpenAI_SendMsg(msg, "english")
    fmt.Println("translated send s=", s)
    s = s[1 : len(s)-1] // remove double quotes


    sendargs[1] = s
    go handleCmd("send", sendargs) // send the msg
    return true
}

// Note: This is not a receiver function of App struct, so we need a context to be
// stored separately. Called dynamically as soon as we receive a msg from the device.
// We will use events from the wails framework to send an event to the frontend
// about the received msg
func ReceiveMsg(source, msg string){
    // filter the msgs which are only received from the source I am interacting with
    // We need to ignore the + sign , therefore use TrimPrefix
    if(source == strings.TrimPrefix(sendargs[0], "+")){
        fmt.Println("original rxd msg:", msg)

    // translate to local language using chatGPT before sending to frontend
        s := OpenAI_SendMsg(msg, translateTo)
        s = s[1: len(s)-1]  // remove double quotes
        fmt.Println("translated recv s=", s)

        rcvargs[0] = s
        runtime.EventsEmit(runtimectx, "rxmsg", rcvargs[0])
    }
}

















// We are now done with sending message to the device.
// frontend  ==> backend  ===> device (send message to the recipient using whatsapp protocol)

// Now the other way, we will receive a message dynamically
// device (user sends the message to me)===> received by backend ==> need to update to frontend


