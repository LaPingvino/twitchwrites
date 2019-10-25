package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"regexp"
	"time"
	"unicode"

	irc "github.com/fluffle/goirc/client"
)

var regex = "^[^ ]+$"

func main() {
	if len(os.Args) < 4 {
		fmt.Printf("%#v", os.Args)
		panic("Cannot login without username/password and channel")
	}
	if len(os.Args) > 4 {
		regex = os.Args[4]
	}
	// Or, create a config and fiddle with it first:
	cfg := irc.NewConfig(os.Args[1])
	cfg.SSL = true
	cfg.SSLConfig = &tls.Config{ServerName: "irc.chat.twitch.tv"}
	cfg.Server = "irc.chat.twitch.tv:6697"
	cfg.Pass = os.Args[2]
	c := irc.Client(cfg)

	// Add handlers to do things here!
	// e.g. join a channel on connect.
	c.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) { conn.Join(os.Args[3]) })
	// And a signal on disconnect
	quit := make(chan bool)
	c.HandleFunc(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	// Tell client to connect.
	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
	}

	var timer = time.Tick(time.Minute)
	var measure map[string]int

	c.HandleFunc(irc.PRIVMSG,
		func(conn *irc.Conn, line *irc.Line) {
			if ok, _ := regexp.MatchString(regex, line.Text()); ok {
				if measure == nil {
					measure = make(map[string]int)
				}
				measure[line.Text()]++
			}
		})
	for range timer {
		var highest int
		chosen := " "
		for k, v := range measure {
			if v > highest {
				chosen = k
				highest = v
			}
		}
		space := " "
		if chosen == " " {
			fmt.Print("_")
			continue
		}
		if !unicode.IsLetter([]rune(chosen)[0]) {
			space = ""
		}
		fmt.Print(space + chosen)
		measure = nil
	}

	// Wait for disconnect
	<-quit
}
