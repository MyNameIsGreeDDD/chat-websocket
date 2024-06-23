import ws from 'k6/ws';
import {check} from 'k6';

export let options = {
    vus: 100,
    iterations: 100,
    duration: '3s',
};

export default function () {
    const url = 'ws://localhost:8080';

    const response = ws.connect(url, null, function (socket) {
        socket.on('open', function () {
            let userId = Math.floor(Math.random() * 1000000) + 1;
            let readUserId = Math.floor(Math.random() * 1000000) + 1;
            let sessionId = generateRandomString(10)

            socket.send(JSON.stringify({
                event: 'Auth',
                user_id: userId,
                session_id: sessionId,
                data: {
                    token: 'asdasd'
                }
            }));

            for (let i = 1; i <= 5; i++) {
                socket.send(
                    JSON.stringify({
                            event: 'MessageRead',
                            user_id: userId,
                            data: {
                                info_for_client: {
                                    chat_id: 1,
                                    message_id: 1,
                                    read_user_id: readUserId,
                                }
                            }
                        }
                    )
                );
            }
        });

        socket.on('message', function (message) {});
        socket.on('close', function () {});
        socket.on('error', function (e) {
        });
    });

    check(response, {'Статус код101': (r) => r && r.status === 101});
}

function generateRandomString(length) {
    let result = '';
    const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    const charactersLength = characters.length;
    for (let i = 0; i < length; i++) {
        result += characters.charAt(Math.floor(Math.random() * charactersLength));
    }
    return result;
}