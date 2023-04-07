package bot

import (
	"fmt"

	"github.com/aigic8/warmlight/pkg/bot/utils"
)

// Commands
const COMMAND_START = "/start"
const COMMAND_SET_ACTIVE_SOURCE = "/setActiveSource"
const COMMAND_ADD_OUTPUT = "/addOutput"

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

func strActiveSourceIsSet(sourceName string, timeoutMinutes int) string {
	return fmt.Sprintf("The source '%s' is set as active source for %d minutes", sourceName, timeoutMinutes)
}

func strSourceDoesExist(sourceName string) string {
	return fmt.Sprintf("Source '%s' does not exist", sourceName)
}

// Outputs
func strOutputNotFound(chatTitle string) string {
	return fmt.Sprintf("No channel with title '%s' was found. Make sure the bot is admin with send message permissions.\nIf it is, make it a normal user and then again an admin.", chatTitle)
}

func strOutputIsAlreadyActive(chatTitle string) string {
	return fmt.Sprintf("Channel '%s' is already active.", chatTitle)
}

func strOutputIsSet(chatTitle string) string {
	return fmt.Sprintf("Channel '%s' is now active!", chatTitle)
}

// IMPORTANT needs support Markdown parsemode
func strQuote(q *utils.Quote) string {
	// TODO find a more efficient way
	var tagsStr string
	for _, tag := range q.Tags {
		tagsStr += "#" + tag + " "
	}
	return fmt.Sprintf("%s\n**%s**\n%s", q.Text, q.MainSource, tagsStr)
}
