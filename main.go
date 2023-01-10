package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"log"
	"os"
	"strings"
)

func printCommandEvents(event *slackevents.AppMentionEvent) {
	//for event := range analyticsChannel {
	fmt.Println("Command Events...")
	fmt.Println(event.User)
	fmt.Println(event.Type)
	fmt.Println(event.EventTimeStamp)
	fmt.Println(event.BotID)
	fmt.Println(event.Text)
	fmt.Println(event.ThreadTimeStamp)
	fmt.Println(event.Channel)
	fmt.Println()
	//}
}

func fileUpload() {
	//File upload bot logic
	os.Setenv("SLACK_BOT_TOKEN", "xoxb-4561617011412-4561659979524-qi3foREnrxr9ag0doTr3kuVV")
	os.Setenv("CHANNEL_ID", "C04GTPU5KDF")
	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))
	channelArray := []string{os.Getenv("CHANNEL_ID")}
	fileArr := []string{"Jotham-Ntekim.pdf"}

	for i := 0; i < len(fileArr); i++ {
		params := slack.FileUploadParameters{
			Channels: channelArray,
			File:     fileArr[i],
		}
		file, err := api.UploadFile(params)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		fmt.Printf("Name: %s, URL: %s\n", file.Name, file.URL)
	}
}

func main() {

	os.Setenv("SLACK_BOT_TOKEN", "xoxb-4561617011412-4561659979524-qi3foREnrxr9ag0doTr3kuVV")
	os.Setenv("SLACK_APP_TOKEN", "xapp-1-A04GHJNHD28-4556239571909-9c59b088979e69c023758e941345fab06516319733b1fcab12bbc9462957165d")

	//Create a new client to slack by giving slack bot token
	//Set debug to true during development mode
	//Add application token option to te client
	client := slack.New(os.Getenv("SLACK_BOT_TOKEN"), slack.OptionDebug(true), slack.OptionAppLevelToken(os.Getenv("SLACK_APP_TOKEN")))

	socket := socketmode.New(
		client,
		socketmode.OptionDebug(true),
		//Add option to set a custom logger
		socketmode.OptionLog(log.New(os.Stdout, "Socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	//ageBot(bot)
	//bot.
	//bot.Command("My yob is <year>", &slacker.CommandDefinition{
	//	Description: "Age calculator",
	//	//Examples: ["My yob is 2020"],
	//	Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
	//		year := request.Param("year")
	//		yob, err := strconv.Atoi(year)
	//		if err != nil {
	//			println("error")
	//		}
	//
	//		age := 2020 - yob
	//
	//		r := fmt.Sprintf("age is %d", age)
	//		response.Reply(r)
	//	},
	//})
	//Context used to cancel goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Making the cancel properly, implementing a graceful shutdown
	go func(ctx context.Context, client *slack.Client, socket *socketmode.Client) {
		//	For loop selects eiter the context cancellation or the events incoming
		for {
			select {
			//	Exit goroutine if context handle is called
			case <-ctx.Done():
				log.Println("Shutting down socketmode listener...")
				return
				//		Case socket events
			case event := <-socket.Events:
				switch event.Type {
				//handle eventAPI events
				case socketmode.EventTypeEventsAPI:
					/*	The event sent on the channel is not the same as
						the eventAPI events, so need to typecast
					*/

					eventsAPI, ok := event.Data.(slackevents.EventsAPIEvent)
					if !ok {
						log.Printf("Could not type cast the event to the EventsAPIEvent: %v\n", event)
						continue
					}

					//sends an acknoledge to the ack server
					socket.Ack(*event.Request)

					//	We have an Evenns API event setup, but this event type can be of many types
					//  so setting up another type switcher

					err := HandleEventMessage(eventsAPI, client)

					if err != nil {
						log.Fatal(err)
					}
				}
			}

		}
	}(ctx, client, socket)
	fileUpload()
	socket.Run()

}

func HandleEventMessage(event slackevents.EventsAPIEvent, client *slack.Client) error {
	switch event.Type {
	//First chaeck if it's a callback event
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent

		//Run another Type switch on the gotten data if it's an AppMentionEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			//go printCommandEvents(ev)
			// Application mentioned since this event is an app mentioned event
			err := HandleAppMentionEventToBot(ev, client)

			if err != nil {
				return err
			}
		}
	default:
		return errors.New("Unsupported event type")
	}
	return nil
}

/*
	This HandleAppMentionEvent logic takes care of
	AppMentionEvent when the bot is mentioned
*/
func HandleAppMentionEventToBot(event *slackevents.AppMentionEvent, client *slack.Client) error {

	//	Grab username based on the ID of the one who mentioned the bot
	user, err := client.GetUserInfo(event.User)
	if err != nil {
		return err
	}

	text := strings.ToLower(event.Text)

	//	create attachment and assign based on the message
	attachment := slack.Attachment{}

	if strings.Contains(text, "hello") || strings.Contains(text, "hi") {
		//Greet the user
		attachment.Text = fmt.Sprintf("Hello %s", user.Name)
		attachment.Color = "#4af030"
	} else if strings.Contains(text, "weather") {
		//	Send a messsage to the user
		attachment.Text = fmt.Sprintf("Weather is sunny today. %s", user.Name)
		attachment.Color = "#4af030"
	} else {
		//	send message to the user
		attachment.Text = fmt.Sprintf("I am good. How are you %s?", user.Name)
	}

	//	Send message to channel
	//  The Channel is available in the event message
	_, _, err = client.PostMessage(event.Channel, slack.MsgOptionAttachments(attachment))

	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	return nil
}
