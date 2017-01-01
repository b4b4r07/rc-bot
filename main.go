package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"

	"github.com/nlopes/slack"
)

var pattern *regexp.Regexp = regexp.MustCompile(`^bot\s+rc\s+(.*)`)

var (
	shell = flag.String("s", os.Getenv("SHELL"), "Specify the shell")
)

func main() {
	flag.Parse()
	api := slack.New(os.Getenv("SLACK_TOKEN"))
	os.Exit(run(api))
}

func run(api *slack.Client) int {
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	var params slack.PostMessageParameters

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
				log.Print("Connected!")

			case *slack.MessageEvent:
				pat := pattern.FindStringSubmatch(ev.Text)
				if len(pat) < 2 {
					break
				}
				switch ev.User {
				case "U0WFNAD1N_":
					result, err := runCommand(pat[1])
					if err == nil {
						params = getPostMessageParameters(result, true)
					} else {
						params = getPostMessageParameters(err.Error(), false)
					}
				default:
					params = getPostMessageParameters(
						fmt.Sprintf("`%s`: permission denied", ev.User),
						false,
					)
				}
				_, _, err := api.PostMessage(ev.Channel, "", params)
				if err != nil {
					log.Print(err)
					return 1
				}

			case *slack.InvalidAuthEvent:
				log.Print("Invalid credentials")
				return 1
			}
		}
	}
}

func runCommand(cmd string) (string, error) {
	out, err := exec.Command(*shell, "-c", cmd).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func getPostMessageParameters(result string, ok bool) slack.PostMessageParameters {
	color := "danger"
	if ok {
		color = "good"
	}

	params := slack.PostMessageParameters{
		Markdown:  true,
		Username:  "rc-bot",
		IconEmoji: ":trollface:",
	}
	params.Attachments = []slack.Attachment{}
	params.Attachments = append(params.Attachments, slack.Attachment{
		Fallback:   "",
		Title:      "",
		Text:       result,
		MarkdownIn: []string{"title", "text", "fields", "fallback"},
		Color:      color,
	})
	return params
}
