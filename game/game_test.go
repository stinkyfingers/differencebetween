package game

import (
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stinkyfingers/differencebetween/api/testingsupport"
	"github.com/stretchr/testify/assert"
)

func TestGetCards(t *testing.T) {
	tests := []struct {
		s3Client      s3iface.S3API
		expectedCards []Card
		expectedError string
	}{
		{
			s3Client: &testingsupport.S3{
				GetObjectOutput: &s3.GetObjectOutput{
					Body: ioutil.NopCloser(strings.NewReader("test\ntest2\n")),
				},
			},
			expectedCards: []Card{Card("test"), Card("test2")},
		},
		{
			s3Client: &testingsupport.S3{
				GetObjectOutput: nil,
				Err:             errors.New("oh no"),
			},
			expectedError: "oh no",
		},
	}
	for _, test := range tests {
		s3Client = test.s3Client
		cards, err := getCards("test_key")
		if test.expectedError != "" {
			assert.EqualError(t, err, test.expectedError)
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, cards, test.expectedCards)
	}
}

func TestCreateRounds(t *testing.T) {
	g := Game{
		RoundsRemaining: 3,
	}
	s3Client = &testingsupport.S3{
		GetObjectOutput: &s3.GetObjectOutput{
			Body: ioutil.NopCloser(strings.NewReader("test1\ntest2\ntest3\ntest4\ntest5\ntest6\n")),
		},
	}

	err := g.createRounds()
	if err != nil {
		t.Error(err)
	}
	for _, round := range g.Rounds {
		for _, card := range round.Setup {
			if card == "" {
				t.Error("expected card")
			}
		}
	}
}

func TestErrCreateRounds(t *testing.T) {
	g := Game{
		RoundsRemaining: 3,
	}
	s3Client = &testingsupport.S3{
		GetObjectOutput: &s3.GetObjectOutput{
			Body: ioutil.NopCloser(strings.NewReader("test1\n")),
		},
	}
	err := g.createRounds()
	if err != ErrTooFewSetups {
		t.Errorf("expected ErrTooFewSetups, got %v", err)
	}
}

func TestDealPunchlines(t *testing.T) {
	tests := []struct {
		game Game
		err  error
	}{
		{
			game: Game{
				Players: []Player{
					{Name: "al"},
					{Name: "bob"},
				},
				Punchlines: []Card{
					Card("1"),
					Card("2"),
					Card("3"),
					Card("4"),
					Card("5"),
					Card("6"),
					Card("7"),
					Card("8"),
					Card("9"),
					Card("10"),
					Card("11"),
					Card("12"),
				},
			},
		},
		{
			game: Game{
				Players: []Player{
					{Name: "al"},
					{Name: "bob"},
				},
				Punchlines: []Card{
					Card("1"),
				},
			},
			err: ErrTooFewPunchlines,
		},
	}
	for _, test := range tests {
		err := test.game.dealPunchlines()
		if test.err != nil {
			assert.EqualError(t, err, test.err.Error())
		} else {
			for _, p := range test.game.Players {
				assert.Equal(t, 6, len(p.Punchlines))
			}
		}
	}
}

func TestLive(t *testing.T) {
	t.Skip("skip live test")
	setups, err := getSetups()
	if err != nil {
		t.Error(err)
	}

	t.Log(setups)
}
