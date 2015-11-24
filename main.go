package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	
	"github.com/syrneus/logapi/logapi"
	
	redis "gopkg.in/redis.v3"
)

var wsaddr = flag.String("wsaddr", "localhost:8080", "service address")
var apiaddr = flag.String("apiaddr", "localhost:8081", "service address")

// CLI flags.
var configFile = flag.String("config", "/etc/logapi/config.yml", "YAML config file path")

// Config for the program.
var config logapi.ConfigValues

// Our logger.
var logger logapi.Logger


// Our open streams.
type stream struct {
	name string
	uuid string 
	messages chan string
}

// How many active streaming connections we have.
var totalconnections   int32

//var []streams Stream

func init() {	
	// Load command line options.
	flag.Parse()

	// Load the YAML config.
	config, _ = logapi.CreateConfig(*configFile)
	
	// Set up the logger based on the configuration.
	if config.Debug.Verbose {
		logger = *logapi.CreateLogger(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
		logger.Info.Println("Debugging mode enabled")
		logger.Info.Printf("Loaded Config: %v", config)
	} else {
		logger = *logapi.CreateLogger(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	}
	
	streams := make([]stream, 100)
	logger.Trace.Printf("We have %d capacity of streams", len(streams))
	
	logger.Trace.Println("Done setting up")
}

/*
 * Main
 */
func main() {
	
	srvaddr := fmt.Sprintf("%s:%d", config.Connection.TCP.Host, config.Connection.TCP.Port)
	logger.Info.Printf("Setting up api server on: %s", srvaddr)
	
	rtr := mux.NewRouter()
	http.Handle("/", rtr)
	rtr.HandleFunc("/hello/{name:[a-z]+}/", helloHandler).Methods("GET")
	rtr.HandleFunc("/stream", wsstream).Methods("GET")
	rtr.HandleFunc("/stream/{name:[a-z]+}/",wsstream).Methods("GET")
	http.ListenAndServe(srvaddr, nil)
	
	//log.Println("Getting ready to run redis")
	//s := Stream{name: "test", uuid: "UUID", messages: make(chan string)}
	//, messages: []string
	//go redisSub(s.name, s.messages)
	
	logger.Trace.Println("Ending main (we should never get here)")
}


/*
 * Simple hello handler 
 * Operates on /hello/{name:[a-z]+/}
 */
func helloHandler(w http.ResponseWriter, r *http.Request) {
  params := mux.Vars(r)
  name := params["name"]
  w.Write([]byte("Hello " + name + "\n"))
}

/*
 * A new websocket request to stream logs
 *
 * Mux params:
 * @param name The stream name.
 *
 */
func wsstream(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
  	name := params["name"]
	  
	logger.Trace.Printf("We got a new websocket connection for stream: %s", name)
	  
	upgrader := websocket.Upgrader{} // use default options
	
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error.Printf("Failed websocket upgrade: %s", err)
		return
	}
	logger.Trace.Println("Done upgrading connection to websockets")
	
	defer c.Close()
	
	// No stream name was passed, so bail
	if (name == "") {
		logger.Trace.Println("No stream name passed, so cutting client off")
		err = c.WriteMessage(websocket.TextMessage, []byte("ERROR: please pass a stream name"))
		return
	}
	
	s := stream{name: name, uuid: "UUID", messages: make(chan string)}
	logger.Trace.Printf("Starting a goroutine to watch for Redis messages on chan %s", name)
	logger.Trace.Printf("Stream obj: #%v", s)
	
	// Start a redis pubsub goroutine.
	go redisSub(s, s.messages)
	
	// Loop over websocket read messages and channel stream messages.
	for {
		newmsg := <-s.messages
		logger.Trace.Printf("Got from chan: %s", newmsg)
		err = c.WriteMessage(websocket.TextMessage, []byte(newmsg))
		if err != nil {
			logger.Error.Printf("Web socket write error: %s", err)
			break
		}
	}
}


/*
 * Subscribe to a stream, listen for all messages and push them to the channel.
 * Meant to be run in a goroutine.
 *
 * @go
 */
func redisSub(s stream, c chan string) {
	logger.Trace.Printf("PubSub subscribing to %s", s.name)
	
	rsrv := fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port)
	logger.Trace.Println("Connecting to redis server: %s", rsrv)
	
	client := redis.NewClient(&redis.Options{
        Addr:     rsrv,
        Password: config.Redis.Auth, // no password set
        DB:       0,  // use default DB
    })
	
	// TODO, handle this failure.
	pubsub, err := client.Subscribe(s.name)
	if err != nil {
		logger.Error.Printf("Redis subscribe error: %s", err)
		os.Exit(1)
	}
	defer pubsub.Close()
	
	logger.Trace.Println("Connected to redis cluster")
	
	/*
	err = client.Publish(stream, "hello").Err()
	if err != nil {
		panic(err)
	}
	*/
	
	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			logger.Trace.Printf("No message in channel? %s", err)
			os.Exit(1)
		}
		
		logger.Trace.Printf("Got redis message: %s => %s", msg.Channel, msg.Payload)
		c <- msg.String()
	}	
}