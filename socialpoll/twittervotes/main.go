package main

import (
	"bufio"
	"encoding/json"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/joeshaw/envdecode"
	"github.com/nsqio/go-nsq"
	"gopkg.in/mgo.v2"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	authClient *oauth.Client
	creds      *oauth.Credentials
	conn       net.Conn
)

type poll struct {
	Options []string
}
type tweet struct {
	Text string
}

var reader io.ReadCloser

func closeConn() {
	if conn != nil {
		conn.Close()
	}
	if reader != nil {
		reader.Close()
	}
}

// We then run a goroutine that blocks waiting for the signal by trying to read from signalChan;
// this is what the <- operator does in this case (it's trying to read from the channel).
// Since we don't care about the type of signal, we don't bother capturing the object returned on the channel.
// Once a signal is received, we set stop to true and close the connection.
// Only when one of the specified signals is sent will the rest of the goroutine code run,
// which is how we are able to perform teardown code before exiting the program.

/*
First, we make the votes channel that we have been talking about throughout this section,
which is a simple channel of strings. Note that it is neither a send (chan<-) nor a receive (<-chan) channel;
in fact, making such channels makes little sense. We then call publishVotes,
passing in the votes channel for it to receive from and capturing the returned stop signal channel as publisherStoppedChan.
Similarly, we call startTwitterStream,
passing in our stopChan function from the beginning
of the main function and the votes channel for it to send to while capturing
the resulting stop signal channel as twitterStoppedChan.

We then start our refresher goroutine,
which immediately enters an infinite for loop before sleeping for a minute and closing the connection via the call to closeConn.
If the stop bool has been set to true (in that previous goroutine), we will break the loop and exit; otherwise,
we will loop around and wait another minute before closing the connection again.
The use of stoplock is important because we have two goroutines that might try to access the stop variable at the same time,
but we want to avoid collisions.

Once the goroutine has started, we block twitterStoppedChan by attempting to read from it.
When successful (which means the signal was sent on stopChan),
we close the votes channel, which will cause the publisher's for...range loop to exit and the publisher itself to stop,
after which the signal will be sent on publisherStoppedChan, which we wait for before exiting.
*/
func main() {
	var ts struct {
		ConsumerKey    string `env:"SP_TWITTER_KEY,required"`
		ConsumerSecret string `env:"SP_TWITTER_SECRET,required"`
		AccessToken    string `env:"SP_TWITTER_ACCESSTOKEN,required"`
		AccessSecret   string `env:"SP_TWITTER_ACCESSSECRET,required"`
	}
	if err := envdecode.Decode(&ts); err != nil {
		log.Fatalln(err)
	}
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				if conn != nil {
					conn.Close()
					conn = nil
				}
				netc, err := net.DialTimeout(netw, addr, 5*time.Second)
				if err != nil {
					return nil, err
				}
				conn = netc
				return netc, nil
			},
		},
	}
	creds = &oauth.Credentials{
		Token:  ts.AccessToken,
		Secret: ts.AccessSecret,
	}
	authClient = &oauth.Client{
		Credentials: oauth.Credentials{
			Token:  ts.ConsumerKey,
			Secret: ts.ConsumerSecret,
		},
	}
	twitterStopChan := make(chan struct{}, 1)
	publisherStopChan := make(chan struct{}, 1)
	stop := false
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		stop = true
		log.Println("Stopping...")
		closeConn()
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	votes := make(chan string) // chan for votes
	go func() {
		pub, _ := nsq.NewProducer("localhost:4150", nsq.NewConfig())
		for vote := range votes {
			pub.Publish("votes", []byte(vote)) // publish vote
		}
		log.Println("Publisher: Stopping")
		pub.Stop()
		log.Println("Publisher: Stopped")
		publisherStopChan <- struct{}{}
	}()
	go func() {
		defer func() {
			twitterStopChan <- struct{}{}
		}()
		for {
			if stop {
				log.Println("Twitter: Stopped")
				return
			}
			time.Sleep(2 * time.Second) // calm
			var options []string
			db, err := mgo.Dial("localhost")
			if err != nil {
				log.Fatalln(err)
			}
			iter := db.DB("ballots").C("polls").Find(nil).Iter()
			var p poll
			for iter.Next(&p) {
				options = append(options, p.Options...)
			}
			iter.Close()
			db.Close()

			hashtags := make([]string, len(options))
			for i := range options {
				hashtags[i] = "#" + strings.ToLower(options[i])
			}

			form := url.Values{"track": {strings.Join(hashtags, ",")}}
			formEnc := form.Encode()

			u, _ := url.Parse("https://stream.twitter.com/1.1/statuses/filter.json")
			req, err := http.NewRequest("POST", u.String(), strings.NewReader(formEnc))
			if err != nil {
				log.Println("creating filter request failed:", err)
			}
			req.Header.Set("Authorization", authClient.AuthorizationHeader(creds, "POST", u, form))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Content-Length", strconv.Itoa(len(formEnc)))

			resp, err := client.Do(req)
			if err != nil {
				log.Println("Error getting response:", err)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				// this is a nice way to see what the error actually is:
				s := bufio.NewScanner(resp.Body)
				s.Scan()
				log.Println(s.Text())
				log.Println(hashtags)
				log.Println("StatusCode =", resp.StatusCode)
				continue
			}

			reader = resp.Body
			decoder := json.NewDecoder(reader)
			for {
				var t tweet
				if err := decoder.Decode(&t); err == nil {
					for _, option := range options {
						if strings.Contains(
							strings.ToLower(t.Text),
							strings.ToLower(option),
						) {
							log.Println("vote:", option)
							votes <- option
						}
					}
				} else {
					break
				}
			}

		}

	}()

	// update by forcing the connection to close
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			closeConn()
			if stop {
				break
			}
		}
	}()

	<-twitterStopChan // important to avoid panic
	close(votes)
	<-publisherStopChan
}
