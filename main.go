/*
Parse the history of a Slack channel for links to music services.
TODO:
	Lookup artists
	Create a chatbot to do stuff with discovered links
*/
package main

import (
	"flag"
	"fmt"
	"github.com/bluele/slack"
	"net/url"
	"os"
	"regexp"
)

const urlExpr string = `[-a-zA-Z0-9:_\+.~#?&//=]{2,256}\.[^@\ ][a-z]{2,12}\b(\/[-a-zA-Z0-9:%_\+.~#?&//=]*)?`

var (
	token       string = ""
	channelName string = ""
)

type Submission interface {
	URL() string // URL of the submission
}

type YouTube struct {
	url *url.URL
}

func (y YouTube) URL() string { return y.url.String() }

type SoundCloud struct {
	url *url.URL
}

func (s SoundCloud) URL() string { return s.url.String() }

func failOnErr(err error) {
	if err != nil {
		fmt.Println("Error: ", err.Error())
		os.Exit(1)
	}
}

func getURL(content string) string {
	return regexp.MustCompile(urlExpr).FindString(content)
}

func main() {
	flag.StringVar(&token, "token", token, "Slack Token")
	flag.StringVar(&channelName, "channel", channelName, "Slack Channel")
	flag.Parse()
	if token == "" || channelName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	api := slack.New(token)
	channel, err := api.FindChannelByName(channelName)
	if err != nil {
		panic(err)
	}
	msgs, err := api.ChannelsHistoryMessages(&slack.ChannelsHistoryOpt{
		Channel: channel.Id,
		Count:   1000,
	})
	if err != nil {
		panic(err)
	}
	for _, msg := range msgs {
		var sub Submission
		if raw := regexp.MustCompile(urlExpr).FindString(msg.Text); raw != "" {
			u, err := url.Parse(raw)
			if err != nil {
				continue
			}
			switch {
			case regexp.MustCompile(".*youtube.*").FindString(u.Host) != "":
				sub = YouTube{url: u}
			case regexp.MustCompile(".*soundcloud.*").FindString(u.Host) != "":
				sub = SoundCloud{url: u}
			default:
				continue
			}
			fmt.Println(sub.URL())
		}
	}
}

func init() {
	token = os.Getenv("SLACK_TOKEN")
	channelName = os.Getenv("SLACK_CHANNEL")
}
