package evtc

import (
	"time"

	"github.com/pkg/errors"
	"golang.org/x/text/language"
)

type EventChain struct {
	agents map[uint64]*Agent
	skills map[uint32]string

	serverTime time.Time
	localTime  time.Time
	timeOffset time.Duration

	ArcDPSVersion string
	BuildID       int
	BossSpecies   int
	BossName      string
	PointOfView   *Agent
	Language      language.Tag
	Events        []Event
	WorldID       uint16
	MapID         uint16
}

func makeEventChain(h header, agents map[uint64]*wrappedAgent, skills map[uint32]string, events []cbtevent1) (*EventChain, error) {
	chain := &EventChain{
		agents: make(map[uint64]*Agent),
		skills: skills,

		ArcDPSVersion: string(h.Date[:]),
	}

	chain.BossSpecies = int(h.Boss)
	for addr, wrapped := range agents {
		if wrapped.speciesID == h.Boss {
			chain.BossName = wrapped.charName
		}

		chain.agents[addr] = &Agent{
			wrapped: wrapped,
			chain:   chain,
		}
	}

	for _, event := range events {
		if e, err := parseEvent(chain, event); err != nil {
			return nil, errors.Wrap(err, "evtc: failed to parse event")
		} else if e != nil {
			chain.Events = append(chain.Events, e)
		}
	}

	return chain, nil
}
