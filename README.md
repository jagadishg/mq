# mq
MQTT CLI - Nice little command line tool for interacting with your MQTT Brokers.

## Connect
mq connect -h wss://livechat.autonom8.com:443 -u jagadish
> Prompt for password
Connected

> onConnect
    > publish / subscribe
> onDisconnect
    > unsubscribe if subscribed

## Connection Status
mq status
Conencted
Host:
Port:
Username:

## Publish
mq publish <topic> <message>

## Subscribe
mq subscribe <topic> <topic> <topic>

    Flow:
    Connect to mq server
    Send subscribe request for requested topic(s)
    Wait for messages

    On MQ Server:
    onConnect
    onSubscribe -> add client to the subscribers list



## Possible Errors for publish and subscribe operations
mqtt broker is not connected
not authorized to publish / subscribe


subscribe
---
topics
if (not already subscribed) {

}

unsubscribe
---