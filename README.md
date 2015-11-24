logapi
======

logapi is a simple web server meant to be the hub for managing log streaming. It is a work in progress ultimately meant to be a downstream consumer of Kinesis streams.

# Architecture
The system runs an API server in Go, powered by Gorilla Toolkit components.  The Gorilla websocket library is used for streaming logs to the client.

### Redis backend
Redis PubSub is used as the backend source of logs to stream to clients.

# Example usage
Start up the API server
```
$ go run main.go
INFO: 2015/11/23 11:52:25 main.go:53: Debugging mode enabled
INFO: 2015/11/23 11:52:25 main.go:54: Loaded Config: {{logapi local} {{true 127.0.0.1 8080}} {127.0.0.1 6379 } {true}}
TRACE: 2015/11/23 11:52:25 main.go:62: Done setting up
INFO: 2015/11/23 11:52:25 main.go:71: Setting up api server on: 127.0.0.1:8080
```

Test hitting the server with a websocket client
```
$ go run client-test.go -path=/stream/mysearch/
Connecting to ws://localhost:8080/stream/mysearch/
URL is: ws://localhost:8080/stream/mysearch/
```

Test publishing a message to Redis
```
$ telnet localhost 6379
Trying ::1...
Connected to localhost.
Escape character is '^]'.
PUBLISH mysearch "testing..."
:1
```

You should see the result in the websocket client
```
recv: Message<mysearch: testing...>
```

# Configuration
The system will look for configuration settings in /etc/logapi/config.yml. An example yaml template is included.

Alternatively, the path to a config file may be passed via the CLI with the -config=/path/to/config.yml option.

### Config options
The config file has options for:
* service.Name - the name of the daemon
* service.Hostname - the hostname to run as
* Connection.TCP.Enabled - whether to run the TCP syslog server
* Connection.TCP.Host - what host to run as
* Connection.TCP.Port - what port to run on
* Redis.Host - the Redis host
* Redis.Port - the Redis port
* Redis.Auth - Redis auth string
* Debug.Verbose - turn on verbose logging

# TODO
* Handle streaming clients dropping. Currently the redis connection stays open and we have multiple stream subscribers
* Integrate with Kinesis to get logs into the Redis server
* API to manage which searches are currently producing which streams
* Auth so not just anyone can stream logs