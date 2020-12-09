var exampleSocket = new WebSocket("ws://localhost:8081/chat");

exampleSocket.onopen = function (event) {
    console.log('websocket open!', event);
};
exampleSocket.onclose = function (event) {
    console.log('websocket closed!', event);
};

exampleSocket.onmessage = receive;

function receive(event) {
    console.log("received event", event);
    var msg = JSON.parse(event.data);
    var body = msg.body;
    
    switch (msg.type) {
        case "count":
            var connections = document.getElementById("connections");
            connections.innerText = body.count.toString();
            break;
        case "message":
            var text = "<div>(" + body.time + ") <b>" + body.name + "</b>: " + body.text + "</div>";
            var chatbox = document.getElementById("chatbox");
            chatbox.contentDocument.write(text);
            // console.log('scrolling to', chatbox.scrollHeight);
            // TODO: Make this more dynamic
            chatbox.contentWindow.scrollBy(0, 18)
            break;
        default:
            break;
    }
}

form = document.getElementById("form")
form.addEventListener('submit', event => {
    event.preventDefault();
    send();
});

function send() {
    var text = document.getElementById("text").value;
    if (text) {
        var msg = {
            type: "message",
            body: {
                text: document.getElementById("text").value,
                name: "admin",
                time: new Date().toLocaleTimeString()
            }
        };

        console.log("sending message", JSON.stringify(msg));
        exampleSocket.send(JSON.stringify(msg));

        document.getElementById("text").value = "";
    }
}