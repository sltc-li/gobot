package handlers

import (
	"errors"
	"fmt"
	"github.com/li-go/gobot/gobot"
	"github.com/li-go/gobot/localrepo"
	"math/rand"
	"regexp"
	"strconv"
	"time"
)

var (
	lunchHelp = "lunch [add|rm|ls|gacha]"

	lunchPattern      = regexp.MustCompile(`^lunch$`)
	lunchAddPattern   = regexp.MustCompile(`^lunch add (.+)$`)
	lunchRmPattern    = regexp.MustCompile(`^lunch rm (.+)$`)
	lunchLsPattern    = regexp.MustCompile(`^lunch ls$`)
	lunchGachaPattern = regexp.MustCompile(`^lunch gacha$`)

	fmtStorageErr = func(err error) error {
		return fmt.Errorf("storage error: %v", err)
	}
)

var lunchHandler = gobot.Handler{
	Name:         "lunch",
	Help:         lunchHelp,
	NeedsMention: false,
	Handleable: func(bot gobot.Bot, msg gobot.Message) bool {
		if msg.Type == gobot.ReplyTo {
			return false
		}
		patterns := []*regexp.Regexp{lunchPattern, lunchAddPattern, lunchRmPattern, lunchLsPattern, lunchGachaPattern}
		for _, p := range patterns {
			if p.MatchString(msg.Text) {
				return true
			}
		}
		return false
	},
	Handle: func(bot gobot.Bot, msg gobot.Message) error {
		if lunchPattern.MatchString(msg.Text) {
			bot.SendMessage("```\n  * "+lunchHelp+"\n```", msg.ChannelID)
			return nil
		}

		ai, err := newLunchAi()
		if err != nil {
			return fmtStorageErr(err)
		}
		defer ai.Close()

		if lunchAddPattern.MatchString(msg.Text) {
			name := lunchAddPattern.FindStringSubmatch(msg.Text)[1]
			r := Restaurant{Name: name}
			if ai.Exists(r) {
				return errors.New("restaurant already exists")
			}
			if err := ai.Add(r); err != nil {
				return fmtStorageErr(err)
			}
			bot.SendMessage("new restaurant added!", msg.ChannelID)
			return nil
		}
		if lunchRmPattern.MatchString(msg.Text) {
			name := lunchRmPattern.FindStringSubmatch(msg.Text)[1]
			r := Restaurant{Name: name}
			if !ai.Exists(r) {
				return errors.New("restaurant doesn't exist")
			}
			if err := ai.Remove(r); err != nil {
				return fmtStorageErr(err)
			}
			bot.SendMessage("restaurant removed!", msg.ChannelID)
			return nil
		}
		if lunchLsPattern.MatchString(msg.Text) {
			restaurants, err := ai.All()
			if err != nil {
				return fmtStorageErr(err)
			}
			s := "```\nRestaurants:\n"
			for i, r := range restaurants {
				s += "  " + strconv.Itoa(i+1) + ". " + r.Name + "\n"
			}
			s += "```"
			bot.SendMessage(s, msg.ChannelID)
			return nil
		}
		if lunchGachaPattern.MatchString(msg.Text) {
			restaurant, err := ai.One()
			if err != nil {
				return fmtStorageErr(err)
			}
			bot.SendMessage("Let's GO *"+restaurant.Name+"* today! :rice:", msg.ChannelID)
			return nil
		}
		return nil
	},
}

type Restaurant struct {
	Name string `db:"name" gorm:"primary_key"`
}

type lunchAi struct {
	repo localrepo.Repository
}

func newLunchAi() (*lunchAi, error) {
	repo, err := localrepo.New()
	if err != nil {
		return nil, err
	}
	if err = repo.Migrate(Restaurant{}); err != nil {
		return nil, err
	}
	return &lunchAi{repo: repo}, nil
}

func (l *lunchAi) Close() error {
	return l.repo.Close()
}

func (l *lunchAi) Add(r Restaurant) error {
	return l.repo.Put(r)
}

func (l *lunchAi) Remove(r Restaurant) error {
	return l.repo.Del(r)
}

func (l *lunchAi) All() ([]Restaurant, error) {
	var rr []Restaurant
	err := l.repo.GetAll(Restaurant{}, &rr)
	return rr, err
}

func (l *lunchAi) One() (*Restaurant, error) {
	var rr []Restaurant
	if err := l.repo.GetAll(Restaurant{}, &rr); err != nil {
		return nil, err
	}
	rand.Seed(time.Now().UnixNano())
	return &rr[rand.Intn(len(rr))], nil
}

func (l *lunchAi) Exists(r Restaurant) bool {
	var nr Restaurant
	if err := l.repo.GetOne(r, &nr); err != nil {
		return false
	}
	return len(nr.Name) > 0
}
