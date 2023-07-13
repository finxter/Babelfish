import './style.css';
import './app.css';

import logo from './assets/images/chatbot.png';
import {InitWhatsApp, SendNumber, SendMsg} from '../wailsjs/go/main/App';

document.querySelector('#app').innerHTML = `
    <img id="logo" class="logo">
        <div><strong>AI Powered Message Translator</strong></div>
            <button class="connect_btn" onclick="connect()">Connect</button>

        <div>
            <label for="language">Select a language:</label>
            <select id="language">
                <option value="german">German</option>
                <option value="hindi">Hindi</option>
                <option value="japanese">Japanese</option>
            </select>
        </div>

        <div class="result" id="result">Enter phone number</div>
            <div class="number-box" id="input">
                <textarea class="input" id="number" type="text" autocomplete="off"></textarea>
                <button class="number_btn" onclick="number_send()">Enter</button>
            </div>
        </div>

        <div class="result" id="result">Enter message below 👇</div>
            <div class="input-box" id="input">
                <textarea class="input" id="msg" type="text" autocomplete="off"></textarea>
                <button class="send_btn" onclick="msg_send()">Send</button>
            </div>
        </div>

        <div>
            <textarea id="dynamic-box"></textarea>
            <button class="clear_btn" onclick="clear_msg()">Clear</button>
        </div>

`;
document.getElementById('logo').src = logo;

let resultElement = document.getElementById("result");
const languageSelectElement = document.getElementById('language');
let numberElement = document.getElementById("number");
let msgElement = document.getElementById("msg");
msgElement.focus();
let dynamicBox = document.getElementById('dynamic-box');


// Connect to whatsapp - link your mobile device to this bot
// You need to scan the QR code on your mobile whatsapp atleast once
const connect = async () => {
    try{
        // Get selected language
        let selectedLanguage = languageSelectElement.value;
        await InitWhatsApp(selectedLanguage);
    }catch (err) {
        console.error(err);
    }
};

const connectButton = document.querySelector('.connect_btn');
connectButton.addEventListener('click', connect);


// Send number with whom you want to send  whatsapp msgs

const number_send = async () => {
    // Get number
    let num = numberElement.value;
    // Check if the input is empty
    if (num === "") return;

    try{
            const result = await SendNumber(num);
        // Clear input box if result is true
        if (result) {
            numberElement.value = "";
        }

    }catch (err) {
        console.error(err);
    }
};

const numberButton = document.querySelector('.number_btn');
numberButton.addEventListener('click', number_send);



// Send the message
const msg_send = async () => {
    // Get name
    let msg = msgElement.value;
    // Check if the input is empty
    if (msg === "") return;
    try{
        const result = await SendMsg(msg );

        // Clear input box if result is true
        if (result) {
            msgElement.value = "";
        }
    }catch (err) {
        console.error(err);
    }
};
const sendButton = document.querySelector('.send_btn');
sendButton.addEventListener('click', msg_send);


// We wait for the events to update the received text/msg dynamically. 
//dynamicBox.value = 'This is the updated text.';
runtime.EventsOn("rxmsg", (msg) =>{
        dynamicBox.value = msg
})

// clear the received msg box.
const clear_msg = async () => {
    try{
        dynamicBox.value =""
    }catch (err) {
        console.error(err);
    }
};

const clearButton = document.querySelector('.clear_btn');
clearButton.addEventListener('click', clear_msg);