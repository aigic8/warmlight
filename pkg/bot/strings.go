package bot

import (
	"fmt"

	"github.com/aigic8/warmlight/internal/db"
	"github.com/aigic8/warmlight/pkg/bot/utils"
	"github.com/go-telegram/bot"
)

// Commands
const COMMAND_START = "/start"
const COMMAND_SET_ACTIVE_SOURCE = "/setactivesource"
const COMMAND_GET_OUTPUTS = "/getoutputs"
const COMMAND_DEACTIVATE_SOURCE = "/deactivatesource"

const strInternalServerErr = "Internal server error happened!\nPlease retry"
const strQuoteAdded = "Quote added"

func strWelcomeToBot(firstName string) string {
	return "Welcome to the bot " + firstName + "!"
}

func strYouAreAlreadyJoined(firstName string) string {
	return "You are already joined " + firstName + "!"
}

func strYourDataIsLost(firstName string) string {
	return "Sorry, it looks like we lost your data" + firstName + "!"
}

// Active Source setting

// TODO
const strMalformedSetActiveSource = "TODO write different examples on how to set active source"
const strSourceTimeoutShouldBeGreaterThanZero = "active source timeout should be greater than zero!"
const strActiveSourceExpired = "active source expired"
const strQuoteAddedButFailedToPublish = "quote is added but failed to publish in channels"
const strNoActiveSource = "You currently have no active sources!"

func strActiveSourceDeactivated(sourceName string) string {
	return "source '" + sourceName + "' deactivated"
}

func strActiveSourceIsSet(sourceName string, timeoutMinutes int) string {
	return fmt.Sprintf("The source '%s' is set as active source for %d minutes", sourceName, timeoutMinutes)
}

func strSourceDoesExist(sourceName string) string {
	return fmt.Sprintf("Source '%s' does not exist", sourceName)
}

// IMPORTANT needs support for Markdown parseMode
func strListOfYourOutputs(outputs []db.Output) string {
	if len(outputs) == 0 {
		return "You have no outputs.\n To add an output you need to set the bot as admin of a channel.\n If you have already done that, please redo it and try again."
	}

	text := ""
	for _, output := range outputs {
		state := "deactive"
		if output.IsActive {
			state = "active"
		}
		text += "*" + bot.EscapeMarkdown(output.Title) + "* - " + state
	}
	return text
}

// IMPORTANT needs support Markdown parseMode
func strQuote(q *utils.Quote) string {
	message := bot.EscapeMarkdown(q.Text)

	if q.MainSource != "" {
		message += "\n" + "*" + bot.EscapeMarkdown(q.MainSource) + "*"
	}

	if q.Tags != nil && len(q.Tags) != 0 {
		// TODO find a more efficient way
		tagsStr := ""
		if q.Tags != nil && len(q.Tags) != 0 {
			for _, tag := range q.Tags {
				tagsStr += "#" + tag + " "
			}
		}

		message += "\n" + bot.EscapeMarkdown(tagsStr)
	}

	return message
}
