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
const InternalServerErr = "Internal server error happened!\nPlease retry"
const QuoteAdded = "Quote added"
const OperationCanceled = "Operation canceled."

func WelcomeToBot(firstName string) string {
	return "Welcome to the bot " + firstName + "!"
}

func YouAreAlreadyJoined(firstName string) string {
	return "You are already joined " + firstName + "!"
}

func YourDataIsLost(firstName string) string {
	return "Sorry, it looks like we lost your data" + firstName + "!"
}

// SOURCES ///////////////////////////////////////////////////////
// TODO: write different examples on how to set active source
const MalformedSetActiveSource = "TODO write different examples on how to set active source"
const SourceTimeoutShouldBeGreaterThanZero = "active source timeout should be greater than zero!"
const ActiveSourceExpired = "active source expired"
const QuoteAddedButFailedToPublish = "quote is added but failed to publish in channels"
const NoActiveSource = "You currently have no active sources!"
const OnlyOneSourceKindFilterIsAllowed = "Currently you can not use more than one filter for sources!"
const SourceNoLongerExists = "Source no longer exists"
const CanceledEditMode = "Operation canceled"
const GoingBackToNormalMode = "There was an error in operation. Canceled the operation."
const MalformedEditSourceText = "TODO edit text was malformed. Please rewrite it."
const MalformedPersonDates = "TODO malformed person dates"

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
	return "source '" + sourceName + "' deactivated"
}

func ActiveSourceIsSet(sourceName string, timeoutMinutes int) string {
	return fmt.Sprintf("The source '%s' is set as active source for %d minutes", sourceName, timeoutMinutes)
}

func SourceDoesNotExist(sourceName string) string {
	return fmt.Sprintf("Source '%s' does not exist", sourceName)
}

func InvalidSourceKind(sourceKind string) string {
	return fmt.Sprintf("Source kind '%s' is not a valid source kind. Valid source kinds are %s.", sourceKind, strings.Join(db.VALID_SOURCE_KINDS, ", "))
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

	return "updated successfully. New source info:\n" + sourceInfoStr, nil
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

	return "Current:\n" + sourceInfo + "To update use bla bla bla. TODO:", nil
}

// IMPORTANT needs support for Markdown parseMode
func ListOfSources(sources []db.Source) string {
	if len(sources) == 0 {
		return "No source was found!"
	}

	text := "Found sources:\n"
	for i, source := range sources {
		text += strconv.Itoa(i+1) + ". *" + bot.EscapeMarkdown(source.Name) + "*" + " \\- " + string(source.Kind) + "\n"
	}

	return text
}

// OUTPUTS ///////////////////////////////////////////////////////
// IMPORTANT needs support for Markdown parseMode
func ListOfYourOutputs(outputs []db.Output) string {
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
const OnlyTheOwnerCanAddNewUsers = "Only the owner of library can add new users!"
const MalformedLibraryToken = "Token is not valid."
const NoLibraryExistsWithToken = "Token is not valid."
const MergeOrDeleteCurrentLibraryData = "Do you want merge your current data or delete it?"
const LibraryTokenExpired = "Token is expired!"
const LibraryNoLongerExistsOPCancled = "Library you wanted to use, no longer exists. Operation canceled."
const ConfirmLibraryChangeCancelAnswer = "Cancel"
const ConfirmLibraryChangeYesAnswer = "Yes, I want use this library."
const UnknownLibraryConfirmationMessage = "Unknwon reply, valid answers are either '" + ConfirmLibraryChangeYesAnswer + "' or '" + ConfirmLibraryChangeCancelAnswer + "'."
const LibraryChangedSuccessfully = "Library changed successfully!"

func YourLibraryToken(token string, lifetimeStr string) string {
	return "Your library token is " + token + ". It will expire in " + lifetimeStr + "."
}

func ConfirmLibraryChange(YesAnswer, NoAnswer string) string {
	return fmt.Sprintf("Are you sure you want to join this library?\nThis action is IRREVERSIBLE. If yes send '%s'. Send '%s' to cancel.", YesAnswer, NoAnswer)
}
