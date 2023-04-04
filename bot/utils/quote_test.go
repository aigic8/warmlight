package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type parseQuoteTestCase struct {
	Name     string
	Text     string
	Quote    *Quote
	MakesErr bool
}

func TestParseQuote(t *testing.T) {
	normalTestCase := parseQuoteTestCase{
		Name: "normal",
		Text: "When communicators are not trying to influence us, their potential to do so is increased.\nsources: The social animal, Elliot Aronson\n #sociology #psychology #influence",
		Quote: &Quote{
			Text:       "When communicators are not trying to influence us, their potential to do so is increased.",
			MainSource: "The social animal",
			Sources:    []string{"Elliot Aronson", "The social animal"},
			Tags:       []string{"sociology", "psychology", "influence"},
		},
		MakesErr: false,
	}

	withoutTextTestCase := parseQuoteTestCase{
		Name:     "withoutText",
		Text:     "\n  \n  \n",
		MakesErr: true,
	}

	withoutSourceTestCase := parseQuoteTestCase{
		Name: "withoutSource",
		Text: "The person who is easiest to brainwash is the person whose beliefes are based on slogans that have never been seriously tested. \n #sociology",
		Quote: &Quote{
			Text:       "The person who is easiest to brainwash is the person whose beliefes are based on slogans that have never been seriously tested.",
			MainSource: "",
			Tags:       []string{"sociology"},
			Sources:    []string{},
		},
		MakesErr: false,
	}

	testCases := []parseQuoteTestCase{normalTestCase, withoutTextTestCase, withoutSourceTestCase}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			q, err := ParseQuote(tc.Text)
			if tc.MakesErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, q.Text, tc.Quote.Text)
			assert.Equal(t, q.MainSource, tc.Quote.MainSource)
			assert.ElementsMatch(t, q.Sources, tc.Quote.Sources)
			assert.ElementsMatch(t, q.Tags, tc.Quote.Tags)
		})
	}

}
