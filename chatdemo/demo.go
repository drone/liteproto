package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/drone/liteproto/liteproto"
	"github.com/drone/liteproto/liteproto/liteprotohttp"
)

/*

To compile the demo:

	go build -o demo github.com/drone/liteproto/chatdemo

To run the demo, open two terminals and in the first one start:

	demo :2001 http://localhost:2002

And in the second start:

	demo :2002 http://localhost:2001

This will run two instances of the application that can talk to each other.
*/

func main() {
	if len(os.Args) != 3 {
		log.Println("usage: demo <local-serve-port> <remote-client-url>")
		log.Println("example: demo :8080 http://localhost:8081")
		return
	}

	local := os.Args[1]
	remote := os.Args[2]

	// Create jobs that will execute locally typed commands and remote calls (the Execer interface).
	jGreet := jobGreet{i: local, they: remote}
	jSay := jobSay{}
	jCount := jobCount{}

	// Create new object for handling HTTP JSON messages
	proto := liteprotohttp.New(
		remote+"/req",
		remote+"/resp",
		true,
		&http.Client{Timeout: 750 * time.Millisecond},
		nil)

	// Register job types
	proto.RegisterWithResponder("greet", jGreet)
	proto.Register("say", jSay)
	proto.RegisterWithResponder("count", jCount)

	// The library needs two different end points for requests and responses.
	mux := http.NewServeMux()
	mux.Handle("/req", proto.HandlerRequest())
	mux.Handle("/resp", proto.HandlerResponse())

	server := &http.Server{Addr: local, Handler: mux}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %s\n", err.Error())
		}

		fmt.Println("Server terminated.")
	}()

	scan := bufio.NewScanner(os.Stdin)

	func() {
		defer server.Shutdown(context.Background())

		fmt.Println("Chat demo. Type commands and hit enter. Available commands are:")
		fmt.Println(" * greet")
		fmt.Println(" * say <message>")
		fmt.Println(" * count <number 1..20>")
		fmt.Println(" * quit")

		for {
			fmt.Println()

			if !scan.Scan() {
				return
			}

			command := strings.TrimSpace(scan.Text())

			var message string
			var r interface {
				commandRun(client liteproto.Client, message string) error
			}

			switch {
			case command == "":
				continue
			case command == "greet":
				r = jGreet
			case strings.HasPrefix(command, "say "):
				message = command[4:]
				r = jSay
			case strings.HasPrefix(command, "count "):
				message = command[6:]
				r = jCount
			case command == "quit":
				return
			default:
				fmt.Printf("Unsupported command: %q\n", command)
				continue
			}

			err := r.commandRun(proto, message)
			if err != nil {
				fmt.Printf("Error processing command %q: %s\n", command, err.Error())
			}
		}
	}()

	wg.Wait()
}

func id() string { return strconv.Itoa(rand.Int()) }

type chat struct {
	Message string `json:"message"`
	Number  int    `json:"number,omitempty"`
}

func newChatMessageJSON(s string) json.RawMessage {
	data, _ := json.Marshal(chat{Message: s})
	return data
}

func newChatMessageWithNumberJSON(s string, number int) json.RawMessage {
	data, _ := json.Marshal(chat{Message: s, Number: number})
	return data
}

func getChat(b json.RawMessage) (*chat, error) {
	m := &chat{}
	err := json.Unmarshal(b, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func printMessageThey(b json.RawMessage) *chat {
	s, err := getChat(b)
	if err != nil || s == nil {
		fmt.Printf("They sent an invalid JSON message: %s\n", err)
		return nil
	}

	if s.Message == "" {
		fmt.Printf("They say nothing...\n")
		return s
	}

	fmt.Printf("They say: %s\n", s.Message)

	return s
}

func printMessageWe(s string) {
	fmt.Printf("We say: %s\n", s)
}

//////////
// Jobs //
//////////

// jobGreet handles greet command and it's reply
type jobGreet struct{ i, they string }

func (g jobGreet) commandRun(client liteproto.Client, _ string) (err error) {
	message := fmt.Sprintf("Greetings %s! I'm %s.", g.they, g.i)

	// This code demonstrates calling a remote server and awaiting a response.
	// Note that there is no timeout. We just trust that the remote server will send back a response, so we wait forever.

	responseCh, stopCh, err := client.CallWithResponse(context.Background(), liteproto.Message{
		MessageID:   id(),
		MessageType: "greet",
		MessageData: newChatMessageJSON(message),
	})
	if err != nil {
		return
	}

	go func() {
		response := <-responseCh
		printMessageThey(response.MessageData)
		close(stopCh)
	}()

	printMessageWe(message)

	return
}

// Exec implements liteproto.ExecerWithResponder interface
func (g jobGreet) Exec(ctx context.Context, job liteproto.Message, client liteproto.ResponderClient) {
	printMessageThey(job.MessageData)

	time.Sleep(time.Second)

	// Demonstrates sending a response back to the caller.

	message := fmt.Sprintf("Hello %s. I'm %s. Nice to meet you.", g.they, g.i)
	err := client.Respond(ctx, newChatMessageJSON(message))
	if err != nil {
		log.Printf("Error. Failed to respond: %s", err.Error())
		return
	}

	printMessageWe(message)
}

// jobGreet handles say command
type jobSay struct{}

func (g jobSay) commandRun(client liteproto.Client, message string) (err error) {

	// Demonstrates calling a remote server without handling responses.
	// If the remote server sends back a response it will be ignored.

	err = client.Call(context.Background(), liteproto.Message{
		MessageID:   id(),
		MessageType: "say",
		MessageData: newChatMessageJSON(message),
	})
	if err != nil {
		return
	}

	printMessageWe(message)

	return
}

// Exec implements liteproto.Execer interface
func (g jobSay) Exec(_ context.Context, job liteproto.Message, _ liteproto.Client) {
	printMessageThey(job.MessageData)
}

// jobCount handles count command
type jobCount struct{}

func (g jobCount) commandRun(client liteproto.Client, message string) (err error) {
	number, err := strconv.Atoi(message)
	if err != nil || number < 1 || number > 20 {
		err = errors.New("please provide a valid integer number from 1 to 20")
		return
	}

	message = fmt.Sprintf("Please count to %d. I'll wait for 10 seconds.", number)
	deadline := time.Now().Add(10 * time.Second)

	// This code demonstrates call deadlines. Channel responseCh will be closed by the library at the deadline.

	responseCh, stopCh, err := client.CallWithDeadline(context.Background(), liteproto.Message{
		MessageID:   id(),
		MessageType: "count",
		MessageData: newChatMessageWithNumberJSON(message, number),
	}, deadline)
	if err != nil {
		return
	}

	go func() {
		defer func() {
			close(stopCh)
			log.Println("Finished waiting for counting.")
		}()

		for {
			response, ok := <-responseCh
			if !ok {
				return
			}
			printMessageThey(response.MessageData)
		}
	}()

	printMessageWe(message)

	return
}

// Exec implements liteproto.ExecerWithResponder interface
func (g jobCount) Exec(ctx context.Context, job liteproto.Message, client liteproto.ResponderClient) {
	c := printMessageThey(job.MessageData)

	for i := 1; i <= c.Number; i++ {
		time.Sleep(time.Second)

		select {
		// The context will be cancelled when deadline expires.
		case <-ctx.Done():
			log.Println("Time for responses expired.")
			return
		default:
		}

		message := fmt.Sprintf("%d...", i)
		err := client.Respond(ctx, newChatMessageJSON(message))
		if err != nil {
			log.Printf("Error. Failed to respond: %s\n", err.Error())
			continue
		}

		printMessageWe(message)
	}

	message := "That's it."
	err := client.Respond(ctx, newChatMessageJSON(message))
	if err != nil {
		fmt.Printf("Error. Failed to respond: %s\n", err.Error())
		return
	}

	printMessageWe(message)
}
