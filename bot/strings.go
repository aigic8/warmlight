package bot

import "fmt"

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

func strActiveSourceIsSet(sourceName string, timeoutMinutes int) string {
	return fmt.Sprintf("The source '%s' is set as active source for %d minutes", sourceName, timeoutMinutes)
}

func strSourceDoesExist(sourceName string) string {
	return fmt.Sprintf("Source '%s' does not exist", sourceName)
}
