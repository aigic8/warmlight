package bot

import (
	"fmt"
	"strconv"

	"github.com/aigic8/warmlight/internal/db"
	"github.com/aigic8/warmlight/pkg/bot/utils"
	"github.com/go-telegram/bot"
)

// Commands
const COMMAND_START = "/start"
const COMMAND_SET_ACTIVE_SOURCE = "/setactivesource"
const COMMAND_GET_OUTPUTS = "/getoutputs"
const COMMAND_DEACTIVATE_SOURCE = "/deactivatesource"
const COMMAND_GET_SOURCES = "/getsources"

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
const strOnlyOneSourceKindFilterIsAllowed = "Currently you can not use more than one filter for sources!"
const strSourceNoLongerExists = "Source no longer exists"

func strActiveSourceDeactivated(sourceName string) string {
	return "source '" + sourceName + "' deactivated"
}

func strActiveSourceIsSet(sourceName string, timeoutMinutes int) string {
	return fmt.Sprintf("The source '%s' is set as active source for %d minutes", sourceName, timeoutMinutes)
}

func strSourceDoesNotExist(sourceName string) string {
	return fmt.Sprintf("Source '%s' does not exist", sourceName)
}

func strSourceInfo(source *db.Source, sourceData any) (string, error) {
	switch source.Kind {
	case db.SourceKindUnknown:
		return source.Name + " (unknown)", nil
	case db.SourceKindBook:
		sd := sourceData.(db.SourceBookData)
		return fmt.Sprintf("%s (book):\ninfo url: %s\nauthor: %s\nauthor url: %s", source.Name, sd.LinkToInfo, sd.Author, sd.LinkToAuthor), nil
	case db.SourceKindPerson:
		sd := sourceData.(db.SourcePersonData)
		var bornOnStr, deathOnStr string
		if !sd.BornOn.IsZero() {
			bornOnStr = strconv.Itoa(sd.BornOn.Year())
		}
		if !sd.DeathOn.IsZero() {
			deathOnStr = strconv.Itoa(sd.DeathOn.Year())
		}
		return fmt.Sprintf("%s (person):\ninfo url: %s\ntitle: %s\nlived in: %s-%s", source.Name, sd.LinkToInfo, sd.Title, bornOnStr, deathOnStr), nil
	case db.SourceKindArticle:
		sd := sourceData.(db.SourceArticleData)
		return fmt.Sprintf("%s (article):\nurl: %s\nauthor: %s\n", source.Name, sd.URL, sd.Author), nil
	default:
		return "", utils.ErrUnknownSourceKind
	}
}

// IMPORTANT needs support for Markdown parseMode
func strListOfSources(sources []db.Source) string {
	if len(sources) == 0 {
		return "No source was found!"
	}

	text := "Found sources:\n"
	for i, source := range sources {
		text += strconv.Itoa(i+1) + ". *" + bot.EscapeMarkdown(source.Name) + "*" + " \\- " + string(source.Kind) + "\n"
	}

	return text
}

// IMPORTANT needs support for Markdown parseMode
func strListOfYourOutputs(outputs []db.Output) string {
	if len(outputs) == 0 {
		return "You have no outputs.\n To add an output you need to set the bot as admin of a channel.\n If you have already done that, please redo it and try again."
	}

	text := "Your outputs are:\n"
	for i, output := range outputs {
		state := "deactive"
		if output.IsActive {
			state = "active"
		}
		text += strconv.Itoa(i+1) + ". *" + bot.EscapeMarkdown(output.Title) + "*" + " \\- " + state + "\n"
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
