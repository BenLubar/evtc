package evtc

import (
	"encoding/json"
	"strconv"
	"strings"
)

type agent struct {
	Addr          uint64
	Prof          uint32
	IsElite       uint32
	Toughness     uint16
	Concentration uint16
	Healing       uint16
	HitboxWidth   uint16
	Condition     uint16
	HitboxHeight  uint16
	Name          [64]byte
	Padding       [4]byte
}

type wrappedAgent struct {
	agent

	// synthesized fields
	firstAware uint64
	lastAware  uint64
	masterAddr uint64
	charName   string
	acctName   string
	subgroup   int
	volatileID uint16
	speciesID  uint16
	instanceID uint16
}

func wrapAgents(agents []agent, events []cbtevent1) map[uint64]*wrappedAgent {
	wrapped := make([]wrappedAgent, len(agents))
	lookup := make(map[uint64]*wrappedAgent, len(agents))
	for i, a := range agents {
		wrapped[i].agent = a
		lookup[a.Addr] = &wrapped[i]
		name := strings.SplitN(string(a.Name[:]), "\x00", 4)
		wrapped[i].charName = name[0]
		wrapped[i].acctName = name[1]
		wrapped[i].subgroup, _ = strconv.Atoi(name[2])
		wrapped[i].lastAware = ^uint64(0)

		if a.IsElite == 0xffffffff && a.Prof>>16 == 0xffff {
			wrapped[i].volatileID = uint16(a.Prof & 0xffff)
		} else if a.IsElite == 0xffffffff && a.Prof>>16 != 0xffff {
			wrapped[i].speciesID = uint16(a.Prof & 0xffff)
		}
	}

	instanceLookup := make(map[uint16][]*wrappedAgent)

	for _, e := range events {
		if e.IsStateChange == 0 {
			if a, ok := lookup[e.SrcAgent]; ok {
				if a.instanceID == 0 {
					instanceLookup[e.SrcInstID] = append(instanceLookup[e.SrcInstID], a)
					a.instanceID = e.SrcInstID
					a.firstAware = e.Time
				}

				a.lastAware = e.Time
			}
		}
	}

	for _, e := range events {
		if e.SrcMasterInstID != 0 {
			if a, ok := lookup[e.SrcAgent]; ok {
				for _, i := range instanceLookup[e.SrcMasterInstID] {
					if i.firstAware < e.Time && i.lastAware > e.Time {
						a.masterAddr = i.Addr
						break
					}
				}
			}
		}
	}

	return lookup
}

type Agent struct {
	chain   *EventChain
	wrapped *wrappedAgent
}

func (a *Agent) MarshalJSON() ([]byte, error) {
	type AgentJSON struct {
		Type   string
		Name   string
		Hitbox struct {
			Width  int
			Height int
		}
		Master *Agent `json:",omitempty"`
	}
	agent := AgentJSON{
		Name:   a.Name(),
		Master: a.Master(),
	}
	agent.Hitbox.Width, agent.Hitbox.Height = a.Hitbox()

	if p, ok := a.Player(); ok {
		agent.Type = "Player"
		type PlayerStats struct {
			Toughness     uint8
			Concentration uint8
			Healing       uint8
			Condition     uint8
		}
		return json.Marshal(struct {
			AgentJSON
			Account    string
			Subgroup   int
			Profession string
			Elite      string `json:",omitempty"`

			Stats PlayerStats
		}{
			AgentJSON:  agent,
			Account:    p.Account,
			Subgroup:   p.Subgroup,
			Profession: p.Profession.String(),
			Elite:      p.EliteSpec.String(),

			Stats: PlayerStats{
				Toughness:     p.Toughness,
				Concentration: p.Concentration,
				Healing:       p.Healing,
				Condition:     p.Condition,
			},
		})
	}

	if n, ok := a.NPC(); ok {
		agent.Type = "NPC"
		type NPCStats struct {
			Toughness     int
			Concentration int
			Healing       int
			Condition     int
		}

		return json.Marshal(struct {
			AgentJSON
			SpeciesID int

			Stats NPCStats
		}{
			AgentJSON: agent,
			SpeciesID: n.SpeciesID,

			Stats: NPCStats{
				Toughness:     n.Toughness,
				Concentration: n.Concentration,
				Healing:       n.Healing,
				Condition:     n.Condition,
			},
		})
	}

	if a.IsGadget() {
		agent.Type = "Gadget"
		return json.Marshal(struct {
			AgentJSON
			ID uint64 `json:",string"`
		}{
			AgentJSON: agent,
			ID:        a.wrapped.Addr,
		})
	}

	panic("unreachable")
}

type PlayerInfo struct {
	Name       string
	Account    string
	Subgroup   int
	Profession ProfessionID
	EliteSpec  EliteSpecID

	Toughness     uint8
	Concentration uint8
	Healing       uint8
	Condition     uint8
}

func (a *Agent) Name() string {
	return a.wrapped.charName
}

func (a *Agent) Master() *Agent {
	return a.chain.agents[a.wrapped.masterAddr]
}

func (a *Agent) Hitbox() (width, height int) {
	return int(a.wrapped.HitboxWidth), int(a.wrapped.HitboxHeight)
}

func (a *Agent) Player() (PlayerInfo, bool) {
	if a.wrapped.IsElite == 0xffffffff {
		return PlayerInfo{}, false
	}

	profession := ProfessionID(a.wrapped.Prof)
	eliteSpec := EliteSpecID(a.wrapped.IsElite)

	if eliteSpec == 1 {
		eliteSpec = hotEliteSpec[profession]
	}

	return PlayerInfo{
		Account:    a.wrapped.acctName,
		Subgroup:   a.wrapped.subgroup,
		Profession: profession,
		EliteSpec:  eliteSpec,

		Toughness:     uint8(a.wrapped.Toughness),
		Concentration: uint8(a.wrapped.Concentration),
		Healing:       uint8(a.wrapped.Healing),
		Condition:     uint8(a.wrapped.Condition),
	}, true
}

type NPCInfo struct {
	SpeciesID int

	Toughness     int
	Concentration int
	Healing       int
	Condition     int
}

func (a *Agent) NPC() (NPCInfo, bool) {
	if a.wrapped.IsElite != 0xffffffff || a.wrapped.Prof>>16 == 0xffff {
		return NPCInfo{}, false
	}

	return NPCInfo{
		SpeciesID: int(a.wrapped.speciesID),

		Toughness:     int(a.wrapped.Toughness),
		Concentration: int(a.wrapped.Concentration),
		Healing:       int(a.wrapped.Healing),
		Condition:     int(a.wrapped.Condition),
	}, true
}

func (a *Agent) IsGadget() bool {
	return a.wrapped.IsElite == 0xffffffff && a.wrapped.Prof>>16 == 0xffff
}
