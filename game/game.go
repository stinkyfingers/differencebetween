package game

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type Game struct {
	ID              int       `json:"id"`
	Players         []Player  `json:"players"`
	Punchlines      []Card    `json:"punchlines"`
	Rounds          []Round   `json:"rounds"`
	RoundsRemaining int       `json:"roundsRemaining"` // zero indexed
	CurrentAction   string    `json:"currentAction"`   // play or vote
	Created         time.Time `json:"-"`
}

type Round struct {
	Setup [2]Card         `json:"setup"`
	Plays map[string]Card `json:"plays"` // Player:Card
	Votes map[string]Card `json:"votes"` // Player:Card
}

type Card string

type Player struct {
	Name       string `json:"name"`
	Punchlines []Card `json:"punchlines"`
}

type Play struct {
	Name      string `json:"name"`
	Punchline Card   `json:"punchline"`
	Vote      Card   `json:"vote"`
	Ping      string `json:"ping"`
}

var (
	s3Client s3iface.S3API

	ErrTooFewSetups     = errors.New("not enough setup cards")
	ErrTooFewPunchlines = errors.New("not enough punchline cards")
	ErrNoGamesAvailable = errors.New("no game ids are available")
	ErrMalformedCSV     = errors.New("malformed csv file")

	games = make(map[int]*Game)
)

const (
	differenceBetweenCardsBucket = "differencebetween"
	setupsKey                    = "setups.txt"
	punchlinesKey                = "punchlines.txt"
	setupsCleanKey               = "setups_clean.txt"
	punchlinesCleanKey           = "punchlines_clean.txt"

	setupsFile     = "setups.csv"
	punchlinesFile = "punchlines.csv"

	region = "us-west-1"

	handSize = 6
	PLAY     = "play"
	VOTE     = "vote"
)

func init() {
	// if os.Getenv("DIFF_ENV") == "local" {
	// 	s3Client = MockS3()
	// 	return
	// }
	var sess *session.Session
	if os.Getenv("DIFF_ENV") == "local" {
		sess = session.Must(session.NewSessionWithOptions(session.Options{
			Profile: "jds",
		}))
	} else {
		var err error
		sess, err = session.NewSession()
		if err != nil {
			log.Fatal(err)
		}
	}
	sess.Config.WithRegion(region)
	s3Client = s3.New(sess)
}

func NewGame(player Player, rounds int, cleanliness string) (*Game, error) {
	punchlines, err := getPunchlines(cleanliness)
	if err != nil {
		return nil, err
	}
	setups, err := getSetups(cleanliness)

	if err != nil {
		return nil, err
	}
	id, err := findID()
	if err != nil {
		return nil, err
	}
	g := &Game{
		ID:              id,
		Players:         []Player{player},
		Punchlines:      punchlines,
		RoundsRemaining: rounds,
		CurrentAction:   PLAY,
	}
	err = g.createRounds(setups)
	if err != nil {
		return nil, err
	}
	err = g.dealPunchlines()
	if err != nil {
		return nil, err
	}
	games[g.ID] = g
	return g, nil
}

func GetGame(id int) (*Game, error) {
	if g, ok := games[id]; !ok {
		return nil, errors.New("game does not exist")
	} else {
		return g, nil
	}
}

func findID() (int, error) {
	maxAttempts := 100
	for i := 0; i < maxAttempts; i++ {
		rand.Seed(time.Now().UnixNano())
		id := rand.Intn(99)
		if game, ok := games[id]; !ok {
			return id, nil
		} else if game.Created.Add(time.Hour * 12).After(time.Now()) {
			games[id] = nil
			return id, nil
		}
	}
	return 0, ErrNoGamesAvailable
}

func (g *Game) createRounds(setups []Card) error {
	setupsNeeded := g.RoundsRemaining * 2
	if setupsNeeded > len(setups) {
		return ErrTooFewSetups
	}
	g.Rounds = make([]Round, g.RoundsRemaining)
	rand.Seed(time.Now().UnixNano())
	setupsMap := make(map[int]Card)
	for i := 0; i < setupsNeeded; i++ {
		index := rand.Intn(len(setups))
		if _, ok := setupsMap[index]; ok {
			i--
			continue
		}
		setupsMap[index] = setups[index]
		setup := g.Rounds[i/2].Setup
		setup[i%2] = setups[index]
		g.Rounds[i/2].Setup = setup
	}
	return nil
}

func getSetups(cleanliness string) ([]Card, error) {
	return getCardsCsv(setupsFile, cleanliness)
}

func getPunchlines(cleanliness string) ([]Card, error) {
	return getCardsCsv(punchlinesFile, cleanliness)
}

func getCardsCsv(key, cleanliness string) ([]Card, error) {
	var cards []Card
	resp, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(differenceBetweenCardsBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(resp.Body)
	for {
		line, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if len(line) != 2 {
			return nil, ErrMalformedCSV
		}
		cleanEnough, err := isCleanEnough(line[1], cleanliness)
		if err != nil {
			return nil, err
		}
		if !cleanEnough {
			continue
		}
		cards = append(cards, Card(strings.TrimSpace(line[0])))
	}
	return cards, nil
}

func isCleanEnough(cardCleanliness, cleanliness string) (bool, error) {
	var ok bool
	var cardRank, rank int
	ranks := map[string]int{
		"G":     0,
		"PG":    1,
		"PG-13": 2,
		"R":     3,
		"X":     4,
	}
	cardRank, ok = ranks[cardCleanliness]
	if !ok {
		return false, ErrMalformedCSV
	}
	rank, ok = ranks[cleanliness]
	if !ok {
		return false, ErrMalformedCSV
	}
	if cardRank <= rank {
		return true, nil
	}
	return false, nil
}

func (g *Game) AddPlayer(player Player) error {
	for _, p := range g.Players {
		if p.Name == player.Name {
			return errors.New("player name already exists")
		}
	}
	g.Players = append(g.Players, player)
	return g.dealPunchlines()
}

func (g *Game) Play(playerName string, card Card) {
	round := g.Rounds[g.RoundsRemaining-1]
	if round.Plays == nil {
		round.Plays = make(map[string]Card)
	}
	round.Plays[playerName] = card
	g.Rounds[g.RoundsRemaining-1] = round
	if len(round.Plays) == len(g.Players) {
		g.CurrentAction = VOTE
	}
	// rm used punchline
	for i, player := range g.Players {
		for j, punchline := range player.Punchlines {
			if punchline == card {
				g.Players[i].Punchlines[j] = g.Players[i].Punchlines[len(g.Players[i].Punchlines)-1]
				g.Players[i].Punchlines = g.Players[i].Punchlines[:len(g.Players[i].Punchlines)-1]
			}
		}
	}
	g.dealPunchlines()
}

func (g *Game) Vote(playerName string, card Card) {
	round := g.Rounds[g.RoundsRemaining-1]
	if round.Votes == nil {
		round.Votes = make(map[string]Card)
	}
	round.Votes[playerName] = card
	g.Rounds[g.RoundsRemaining-1] = round
	if len(round.Votes) == len(g.Players) {
		g.RoundsRemaining--
		g.dealPunchlines()
		g.CurrentAction = PLAY
	}
}

func (g *Game) dealPunchlines() error {
	rand.Seed(time.Now().UnixNano())
	for playerIndex := range g.Players {
		cardsNeeded := handSize - len(g.Players[playerIndex].Punchlines)
		if cardsNeeded > len(g.Punchlines) {
			return ErrTooFewPunchlines
		}
		for i := 0; i < cardsNeeded; i++ {
			index := rand.Intn(len(g.Punchlines))
			card := g.Punchlines[index]
			g.Punchlines[index] = g.Punchlines[len(g.Punchlines)-1]
			g.Punchlines = g.Punchlines[:len(g.Punchlines)-1]
			punchlines := g.Players[playerIndex].Punchlines
			punchlines = append(punchlines, card)
			g.Players[playerIndex].Punchlines = punchlines
		}
	}
	return nil
}
