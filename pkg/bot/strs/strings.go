package strs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aigic8/warmlight/internal/db"
	"github.com/aigic8/warmlight/pkg/bot/utils"
	"github.com/go-telegram/bot"
)

// COMMANDS ///////////////////////////////////////////////////////
const COMMAND_START = "/start"
const COMMAND_SET_ACTIVE_SOURCE = "/setactivesource"
const COMMAND_GET_OUTPUTS = "/getoutputs"
const COMMAND_DEACTIVATE_SOURCE = "/deactivatesource"
const COMMAND_GET_SOURCES = "/getsources"
const COMMAND_GET_LIBRARY_TOKEN = "/getlibtoken"
const COMMAND_SET_LIBRARY_TOKEN = "/setlibtoken"

// COMMON STRINGS ////////////////////////////////////////////////
const InternalServerErr = "❌ Something unexpected happened.Text me (@aigic8) if the issue persists."
const QuoteAdded = "✅ Quote added"
const OperationCanceled = "✅ Operation canceled"

func WelcomeToBot(firstName string) string {
	return fmt.Sprintf(`👋 Welcome %s.
Some notes on the bot:
- This is a project I did for fun, and I don't do regular maintenance for it.
- I am not responsible if you loose any valuable information.
- I am not responsible for the content you put in the bot.
This project is open source under MIT LICENSE (the source code is available) so you can easily download the source and create a clone of this bot.
http://github.com/aigic8/warmlight
Feel free to text me if you have any questions/ideas for the bot. (@aigic8)`, firstName)
}

func YouAreAlreadyJoined(firstName string) string {
	return "Joining this bot is like seeing a good movie for the first time, you can only experience it for once. 😌"
}

func YourDataIsLost(firstName string) string {
	return "Looks like your data is lost.\nFeel free to text me (@aigic8) if you've lost any valuable information."
}

// SOURCES ///////////////////////////////////////////////////////
func MalformedSetActiveSource(defaultTimeMins int) string {
	return fmt.Sprintf(`Couldn't understand what you mean. 🤔
To use %s properly you should follow this format:
%s [sourceName], [timeInMins]?
like:
%s Animal Farm, 80
This will set "Animal Farm" source active for 80 minutes
The time parameter is optional. So if you send:
%s Animal Farm
This will make "Animal Farm" source active for a default duration, which is %d minutes.
`, COMMAND_SET_ACTIVE_SOURCE, COMMAND_SET_ACTIVE_SOURCE, COMMAND_SET_ACTIVE_SOURCE, COMMAND_SET_ACTIVE_SOURCE, defaultTimeMins)
}

const SourceTimeoutShouldBeGreaterThanZero = "Active source timeout should greater than zero. 🧐"
const ActiveSourceExpired = "✅ Active source expired."
const QuoteAddedButFailedToPublish = "❌ Quote is added, but failed to publish it to outputs."
const NoActiveSource = "Currently you have no active source. 😊"
const OnlyOneSourceKindFilterIsAllowed = "You can only filter sources based on one source kind. 🧐"
const SourceNoLongerExists = "❌ Source no longer exists."
const GoingBackToNormalMode = "❌ There was an error in operation. Operation is canceled and you went back to normal state."

var MalformedEditSourceText = fmt.Sprintf(`Couldn't understand what you mean. 🤔
To edit the source properly, you should use this format:
[option1]: [value1]
[option2]: [value2]
...
For example for a book:
%s: book
%s: https://en.wikipedia.org/wiki/Animal_Farm
%s: George Orwell
%s: https://en.wikipedia.org/wiki/George_Orwell
You can set the source kind with option named '%s'. Based on source kind you will have different options:
book: %s, %s, %s
person: %s, %s, %s
article: %s, %s
unknown: [HAVE NO OPTIONS]
You can also send '%s' to cancel the operation.`, SOURCE_KIND, SOURCE_BOOK_INFO_URL, SOURCE_BOOK_AUTHOR, SOURCE_BOOK_AUTHOR_URL, SOURCE_KIND, SOURCE_BOOK_INFO_URL, SOURCE_BOOK_AUTHOR, SOURCE_BOOK_AUTHOR_URL, SOURCE_PERSON_INFO_URL, SOURCE_PERSON_LIVED_IN, SOURCE_PERSON_TITLE, SOURCE_ARTICLE_URL, SOURCE_ARTICLE_AUTHOR, ConfirmLibraryChangeCancelAnswer)

var MalformedPersonDates = fmt.Sprintf(`Malformed value for '%s'. The correct format is:
%s: 1960-2000`, SOURCE_PERSON_LIVED_IN, SOURCE_PERSON_LIVED_IN)

const SOURCE_NAME = "name"
const SOURCE_KIND = "kind"

const SOURCE_BOOK_AUTHOR = "author"
const SOURCE_BOOK_INFO_URL = "info url"
const SOURCE_BOOK_AUTHOR_URL = "author url"

const SOURCE_PERSON_INFO_URL = "info url"
const SOURCE_PERSON_TITLE = "title"
const SOURCE_PERSON_LIVED_IN = "lived in"

const SOURCE_ARTICLE_AUTHOR = "author"
const SOURCE_ARTICLE_URL = "url"

func ActiveSourceDeactivated(sourceName string) string {
	return "✅ Source '" + sourceName + "' deactivated."
}

func ActiveSourceIsSet(sourceName string, timeoutMinutes int) string {
	return fmt.Sprintf("✅ Source '%s' is activated for %d minutes.", sourceName, timeoutMinutes)
}

func SourceDoesNotExist(sourceName string) string {
	return fmt.Sprintf("Source '%s' does not exist. 🤔", sourceName)
}

func InvalidSourceKind(sourceKind string) string {
	return fmt.Sprintf("Source kind '%s' is not a valid source kind.🤔\nValid source kinds are %s.", sourceKind, strings.Join(db.VALID_SOURCE_KINDS, ", "))
}

func UpdatedSource(newSource *db.Source) (string, error) {
	data, err := utils.ParseSourceData(newSource.Kind, newSource.Data)
	if err != nil {
		return "", err
	}

	sourceInfoStr, err := SourceInfo(newSource, data)
	if err != nil {
		return "", err
	}

	return "✅ Updated successfully. New source info:\n" + sourceInfoStr, nil
}

func SourceInfo(source *db.Source, sourceData any) (string, error) {
	switch source.Kind {
	case db.SourceKindUnknown:
		return source.Name + " (unknown)", nil
	case db.SourceKindBook:
		sd := sourceData.(db.SourceBookData)
		return fmt.Sprintf("%s (book):\n%s: %s\n%s: %s\n%s: %s", source.Name, SOURCE_BOOK_INFO_URL, sd.LinkToInfo, SOURCE_BOOK_AUTHOR, sd.Author, SOURCE_BOOK_AUTHOR_URL, sd.LinkToAuthor), nil
	case db.SourceKindPerson:
		sd := sourceData.(db.SourcePersonData)
		var bornOnStr, deathOnStr string
		if !sd.BornOn.IsZero() {
			bornOnStr = strconv.Itoa(sd.BornOn.Year())
		}
		if !sd.DeathOn.IsZero() {
			deathOnStr = strconv.Itoa(sd.DeathOn.Year())
		}
		return fmt.Sprintf("%s (person):\n%s: %s\n%s: %s\n%s: %s-%s", source.Name, SOURCE_PERSON_INFO_URL, sd.LinkToInfo, SOURCE_PERSON_TITLE, sd.Title, SOURCE_PERSON_LIVED_IN, bornOnStr, deathOnStr), nil
	case db.SourceKindArticle:
		sd := sourceData.(db.SourceArticleData)
		return fmt.Sprintf("%s (article):\n%s: %s\n%s: %s\n", source.Name, SOURCE_ARTICLE_URL, sd.URL, SOURCE_ARTICLE_URL, sd.Author), nil
	default:
		return "", utils.ErrUnknownSourceKind
	}
}

func EditSource(source *db.Source, sourceData any) (string, error) {
	sourceInfo, err := SourceInfo(source, sourceData)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`Current source info:
%s
Send a message in this format to edit source info:
[option1]: [value1]
[option2]: [value2]
...
For example for a book:
%s: book
%s: https://en.wikipedia.org/wiki/Animal_Farm
%s: George Orwell
%s: https://en.wikipedia.org/wiki/George_Orwell
You can set the source kind with option named '%s'. Based on source kind you will have different options:
book: %s, %s, %s
person: %s, %s, %s
article: %s, %s
unknown: [HAVE NO OPTIONS]
You can also send '%s' to cancel the operation.`, sourceInfo, SOURCE_KIND, SOURCE_BOOK_INFO_URL, SOURCE_BOOK_AUTHOR, SOURCE_BOOK_AUTHOR_URL, SOURCE_KIND, SOURCE_BOOK_INFO_URL, SOURCE_BOOK_AUTHOR, SOURCE_BOOK_AUTHOR_URL, SOURCE_PERSON_INFO_URL, SOURCE_PERSON_LIVED_IN, SOURCE_PERSON_TITLE, SOURCE_ARTICLE_URL, SOURCE_ARTICLE_AUTHOR, ConfirmLibraryChangeCancelAnswer), nil

}

// IMPORTANT needs support for Markdown parseMode
func ListOfSources(sources []db.Source) string {
	if len(sources) == 0 {
		return bot.EscapeMarkdown("No source was found. 😕")
	}

	text := "✅ Found sources:\n"
	for i, source := range sources {
		text += strconv.Itoa(i+1) + bot.EscapeMarkdown(".") + " *" + bot.EscapeMarkdown(source.Name) + "*" + bot.EscapeMarkdown(" - ") + string(source.Kind) + "\n"
	}

	return text
}

// OUTPUTS ///////////////////////////////////////////////////////
// IMPORTANT needs support for Markdown parseMode
func ListOfYourOutputs(outputs []db.Output) string {
	if len(outputs) == 0 {
		return bot.EscapeMarkdown(`You have no outputs.
To add an output you need to set the bot as admin of a channel.
If you have already done that, please remove the bot and make it as admin again. If the issue persists, feel free to contact me (@aigic8).`)
	}

	text := "✅Your outputs are:\n"
	for i, output := range outputs {
		state := "deactive"
		if output.IsActive {
			state = "active"
		}
		text += strconv.Itoa(i+1) + bot.EscapeMarkdown(".") + " *" + bot.EscapeMarkdown(output.Title) + "*" + bot.EscapeMarkdown(" - "+state) + "\n"
	}
	return text
}

// QUOTES ////////////////////////////////////////////////////////
// IMPORTANT needs support Markdown parseMode
func Quote(q *utils.Quote) string {
	message := bot.EscapeMarkdown(q.Text)

	if q.MainSource != "" {
		message += "\n" + "*" + bot.EscapeMarkdown(q.MainSource) + "*"
	}

	if q.Tags != nil && len(q.Tags) != 0 {
		// TODO: find a more efficient way
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

// LIBRARIES /////////////////////////////////////////////////////
const OnlyTheOwnerCanAddNewUsers = "❌ Only the owner of library can add new users. (the first person who have created the library)"

var MalformedLibraryToken = fmt.Sprintf(`Couldn't understand what you mean. 🤔
To use '%s' command properly, use it in this format:
%s [libraryToken]`, COMMAND_SET_LIBRARY_TOKEN, COMMAND_SET_LIBRARY_TOKEN)

const NoLibraryExistsWithToken = "❌ Library token is not valid."
const MergeOrDeleteCurrentLibraryData = `Do you want merge your current data or delete it?
If you merge, you current data will be added to the library your joining.
If you delete, your current data will PERMANENTLY deleted.`
const LibraryTokenExpired = "❌ Library Token is expired.\nAsk the owner for a new token."
const LibraryNoLongerExistsOPCanceled = "❌ Library you wanted to join, no longer exists."
const ConfirmLibraryChangeCancelAnswer = "cancel"
const ConfirmLibraryChangeYesAnswer = "Yes, I want use this library."
const UnknownLibraryConfirmationMessage = "Couldn't understand what you mean.\nValid answers are either '" + ConfirmLibraryChangeYesAnswer + "' or '" + ConfirmLibraryChangeCancelAnswer + "'."
const LibraryChangedSuccessfully = "✅ Library changed successfully."

func YourLibraryToken(token string, lifetimeStr string) string {
	return "✅ Your library token is '" + token + "'. It will expire in " + lifetimeStr + ".\nOnly share it with PEOPLE YOU TRUST."
}

func ConfirmLibraryChange(YesAnswer, NoAnswer string) string {
	return fmt.Sprintf("Are you sure you want to join this library?\nThis action is IRREVERSIBLE. If yes send '%s'. Send '%s' to cancel.", YesAnswer, NoAnswer)
}
