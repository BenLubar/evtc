package evtc

type cbtevent0 struct {
	Time            uint64 /* timegettime() at time of event */
	SrcAgent        uint64 /* unique identifier */
	DstAgent        uint64 /* unique identifier */
	Value           int32  /* event-specific */
	BuffDmg         int32  /* estimated buff damage. zero on application event */
	OverstackValue  uint16 /* estimated overwritten stack duration for buff application */
	SkillID         uint16 /* skill id */
	SrcInstID       uint16 /* agent map instance id */
	DstInstID       uint16 /* agent map instance id */
	SrcMasterInstID uint16 /* master source agent map instance id if source is a minion/pet */
	_               uint8  /* internal tracking. garbage */
	_               uint8  /* internal tracking. garbage */
	_               uint8  /* internal tracking. garbage */
	_               uint8  /* internal tracking. garbage */
	_               uint8  /* internal tracking. garbage */
	_               uint8  /* internal tracking. garbage */
	_               uint8  /* internal tracking. garbage */
	_               uint8  /* internal tracking. garbage */
	_               uint8  /* internal tracking. garbage */
	Iff             uint8  /* from iff enum */
	Buff            uint8  /* buff application, removal, or damage event */
	Result          uint8  /* from cbtresult enum */
	IsActivation    uint8  /* from cbtactivation enum */
	IsBuffRemove    uint8  /* buff removed. src=relevant, dst=caused it (for strips/cleanses). from cbtr enum */
	IsNinety        uint8  /* source agent health was over 90% */
	IsFifty         uint8  /* target agent health was under 50% */
	IsMoving        uint8  /* source agent was moving */
	IsStateChange   uint8  /* from cbtstatechange enum */
	IsFlanking      uint8  /* target agent was not facing source */
	IsShields       uint8  /* all or part damage was vs barrier/shield */
	IsOffCycle      uint8  /* zero if buff dmg happened during tick, non-zero otherwise */
	_               uint8  /* internal tracking. garbage */
}

func convert0(e cbtevent0) cbtevent1 {
	return cbtevent1{
		Time:            e.Time,
		SrcAgent:        e.SrcAgent,
		DstAgent:        e.DstAgent,
		Value:           e.Value,
		BuffDmg:         e.BuffDmg,
		OverstackValue:  uint32(e.OverstackValue),
		SkillID:         uint32(e.SkillID),
		SrcInstID:       e.SrcInstID,
		DstInstID:       e.DstInstID,
		SrcMasterInstID: e.SrcMasterInstID,
		DstMasterInstID: 0,
		Iff:             e.Iff,
		Buff:            e.Buff,
		Result:          e.Result,
		IsActivation:    e.IsActivation,
		IsBuffRemove:    e.IsBuffRemove,
		IsNinety:        e.IsNinety,
		IsFifty:         e.IsFifty,
		IsMoving:        e.IsMoving,
		IsStateChange:   e.IsStateChange,
		IsFlanking:      e.IsFlanking,
		IsShields:       e.IsShields,
		IsOffCycle:      e.IsOffCycle,
	}
}

type cbtevent1 struct {
	Time            uint64
	SrcAgent        uint64
	DstAgent        uint64
	Value           int32
	BuffDmg         int32
	OverstackValue  uint32
	SkillID         uint32
	SrcInstID       uint16
	DstInstID       uint16
	SrcMasterInstID uint16
	DstMasterInstID uint16
	Iff             uint8
	Buff            uint8
	Result          uint8
	IsActivation    uint8
	IsBuffRemove    uint8
	IsNinety        uint8
	IsFifty         uint8
	IsMoving        uint8
	IsStateChange   uint8
	IsFlanking      uint8
	IsShields       uint8
	IsOffCycle      uint8
	Pad61_64        uint32
}
