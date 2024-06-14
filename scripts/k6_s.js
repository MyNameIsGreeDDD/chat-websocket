import ws from 'k6/ws';
import {check} from 'k6';

let userCounter = 0; // Переменная для хранения текущего user_id

export default function () {
    const url = 'ws://localhost:8080';
    const params = {tags: {my_tag: 'hello'}};

    const response = ws.connect(url, params, function (socket) {
        socket.on('open', function ()  {
            userCounter++
            console.log(`connected with user_id: ${userCounter}`);

            // Send Auth event
            const authEvent = {
                event: 'Auth',
                user_id: userCounter,
                data: {
                    user_id: userCounter
                }
            };
            const jsonData = JSON.stringify(authEvent);

            socket.send(jsonData);
            console.log('Auth event sent:', JSON.stringify(authEvent));

            // // Send multiple MessageRead events
            // for (let i = 1; i <= 5; i++) {
            //     const messageReadEvent = {
            //         event: 'MessageRead',
            //         user_id: user_id,
            //         data: {
            //             chat_id: 1,
            //             message_id: i,
            //             session_id: 'asdasd'
            //         }
            //     };
            //     socket.send(JSON.stringify(messageReadEvent));
            //     console.log(`MessageRead event ${i} sent for user_id ${user_id}:`, JSON.stringify(messageReadEvent));
            // }
        });

        socket.on('message', function (message) {
            console.log('Message received: ');
        });

        socket.on('close', function () {
            console.log('disconnected');
        });

        socket.on('error', function (e) {
            console.log('An unexpected error occurred: ', e.error());
        });

        // Automatically close the connection after a short delay
        socket.setTimeout(function () {
            socket.close();
        }, 10000); // Close the connection after 10 seconds
    });

    check(response, {'status is 101': (r) => r && r === 101});
}
