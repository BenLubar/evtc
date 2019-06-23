package evtc

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
)

// ProfessionID is the ID of a Guild Wars 2 profession.
type ProfessionID int

const (
	Guardian     ProfessionID = 1
	Warrior      ProfessionID = 2
	Engineer     ProfessionID = 3
	Ranger       ProfessionID = 4
	Thief        ProfessionID = 5
	Elementalist ProfessionID = 6
	Mesmer       ProfessionID = 7
	Necromancer  ProfessionID = 8
	Revenant     ProfessionID = 9
)

func (id ProfessionID) String() string {
	switch id {
	case Guardian:
		return "Guardian"
	case Warrior:
		return "Warrior"
	case Engineer:
		return "Engineer"
	case Ranger:
		return "Ranger"
	case Thief:
		return "Thief"
	case Elementalist:
		return "Elementalist"
	case Mesmer:
		return "Mesmer"
	case Necromancer:
		return "Necromancer"
	case Revenant:
		return "Revenant"
	default:
		return strconv.Itoa(int(id))
	}
}

// EliteSpecID is the ID of a Guild Wars 2 elite specialization.
type EliteSpecID int

const (
	Druid        EliteSpecID = 5
	Daredevil    EliteSpecID = 7
	Berserker    EliteSpecID = 18
	Dragonhunter EliteSpecID = 27
	Reaper       EliteSpecID = 34
	Chronomancer EliteSpecID = 40
	Scrapper     EliteSpecID = 43
	Tempest      EliteSpecID = 48
	Herald       EliteSpecID = 52
	Soulbeast    EliteSpecID = 55
	Weaver       EliteSpecID = 56
	Holosmith    EliteSpecID = 57
	Deadeye      EliteSpecID = 58
	Mirage       EliteSpecID = 59
	Scourge      EliteSpecID = 60
	Spellbreaker EliteSpecID = 61
	Firebrand    EliteSpecID = 62
	Renegade     EliteSpecID = 63
)

var hotEliteSpec = map[ProfessionID]EliteSpecID{
	Guardian:     Dragonhunter,
	Warrior:      Berserker,
	Engineer:     Scrapper,
	Ranger:       Druid,
	Thief:        Daredevil,
	Elementalist: Tempest,
	Mesmer:       Chronomancer,
	Necromancer:  Reaper,
	Revenant:     Herald,
}

var eliteSpecName = map[EliteSpecID]string{
	Druid:        "Druid",
	Daredevil:    "Daredevil",
	Berserker:    "Berserker",
	Dragonhunter: "Dragonhunter",
	Reaper:       "Reaper",
	Chronomancer: "Chronomancer",
	Scrapper:     "Scrapper",
	Tempest:      "Tempest",
	Herald:       "Herald",
	Soulbeast:    "Soulbeast",
	Weaver:       "Weaver",
	Holosmith:    "Holosmith",
	Deadeye:      "Deadeye",
	Mirage:       "Mirage",
	Scourge:      "Scourge",
	Spellbreaker: "Spellbreaker",
	Firebrand:    "Firebrand",
	Renegade:     "Renegade",
}

var apiEliteSpecName map[EliteSpecID]string

func populateAPIEliteSpecs() {
	resp, err := http.Get("https://api.guildwars2.com/v2/specializations?ids=all&lang=en&v=2019-06-17")
	if err != nil {
		// can't do anything about it
		return
	}
	defer resp.Body.Close()

	var specs []struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Elite bool   `json:"elite"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&specs); err != nil {
		// can't do anything about it
		return
	}

	apiEliteSpecName = make(map[EliteSpecID]string)
	for _, s := range specs {
		if s.Elite {
			apiEliteSpecName[EliteSpecID(s.ID)] = s.Name
		}
	}
}

var populateAPIEliteSpecsOnce sync.Once

func (id EliteSpecID) String() string {
	if id == 0 {
		return ""
	}

	if name, ok := eliteSpecName[id]; ok {
		return name
	}

	populateAPIEliteSpecsOnce.Do(populateAPIEliteSpecs)

	if name, ok := apiEliteSpecName[id]; ok {
		return name
	}

	return strconv.Itoa(int(id))
}
