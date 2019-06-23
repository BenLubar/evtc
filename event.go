package evtc

import (
	"encoding/binary"
	"math"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/text/language"
)

// Event is the base interface all events implement.
type Event interface {
	// Time returns the local and server time of this event.
	// This is computed from an offset from the log start timestamp.
	Time() (local, server time.Time)

	// SourceAgent returns the agent that caused this event.
	// This method may return nil.
	SourceAgent() *Agent
}

// BaseEvent is the shared implementation of Event.
type BaseEvent struct {
	Type       string
	LocalTime  time.Time
	ServerTime time.Time
	Source     *Agent
}

// Time implements Event.
func (e *BaseEvent) Time() (local, server time.Time) {
	return e.LocalTime, e.ServerTime
}

// SourceAgent implements Event.
func (e *BaseEvent) SourceAgent() *Agent {
	return e.Source
}

// SkillEvent is the base interface for events related to skills.
type SkillEvent interface {
	Event

	// Skill returns the ID and name of the skill that was used.
	// For unknown events, name may be empty or the same as the ID.
	Skill() (id int, name string)
}

// CombatEvent is the base interface all events with targets implement.
type CombatEvent interface {
	SkillEvent

	// TargetAgent returns the agent directly affected by this event.
	// This method may return nil.
	TargetAgent() *Agent

	IsFriend() bool
	IsFoe() bool
	IsNinety() bool
	IsFifty() bool
	IsMoving() bool
	IsFlanking() bool
}

// CommonEvent is the shared implementation of CombatEvent.
type CommonEvent struct {
	BaseEvent
	Target    *Agent
	SkillID   int
	SkillName string

	Friend   bool
	Foe      bool
	Ninety   bool
	Fifty    bool
	Moving   bool
	Flanking bool
}

// TargetAgent implements CombatEvent.
func (e *CommonEvent) TargetAgent() *Agent {
	return e.Target
}

// Skill implements CombatEvent.
func (e *CommonEvent) Skill() (id int, name string) {
	return e.SkillID, e.SkillName
}

// IsFriend implements CombatEvent.
func (e *CommonEvent) IsFriend() bool { return e.Friend }

// IsFoe implements CombatEvent.
func (e *CommonEvent) IsFoe() bool { return e.Foe }

// IsNinety implements CombatEvent.
func (e *CommonEvent) IsNinety() bool { return e.Ninety }

// IsFifty implements CombatEvent.
func (e *CommonEvent) IsFifty() bool { return e.Fifty }

// IsMoving implements CombatEvent.
func (e *CommonEvent) IsMoving() bool { return e.Moving }

// IsFlanking implements CombatEvent.
func (e *CommonEvent) IsFlanking() bool { return e.Flanking }

func makeBaseEvent(typ string, chain *EventChain, event cbtevent1) BaseEvent {
	offset := time.Duration(event.Time)*time.Millisecond + chain.timeOffset

	return BaseEvent{
		Type:       typ,
		LocalTime:  chain.localTime.Add(offset),
		ServerTime: chain.localTime.Add(offset),
		Source:     chain.agents[event.SrcAgent],
	}
}

func makeCommonEvent(typ string, chain *EventChain, event cbtevent1) CommonEvent {
	return CommonEvent{
		BaseEvent: makeBaseEvent(typ, chain, event),
		Target:    chain.agents[event.DstAgent],
		SkillID:   int(event.SkillID),
		SkillName: chain.skills[event.SkillID],
		Ninety:    event.IsNinety != 0,
		Fifty:     event.IsFifty != 0,
		Moving:    event.IsMoving != 0,
		Flanking:  event.IsFlanking != 0,
		Friend:    event.Iff == 0,
		Foe:       event.Iff == 1,
	}
}

func parseEvent(chain *EventChain, event cbtevent1) (Event, error) {
	switch {
	case event.IsStateChange != 0:
		return parseStateChangeEvent(chain, event)
	case event.IsActivation != 0:
		return parseActivationEvent(chain, event)
	case event.IsBuffRemove != 0:
		return parseBuffRemoveEvent(chain, event)
	case event.BuffDmg != 0:
		return parseBuffDamageEvent(chain, event)
	case event.Buff != 0:
		return parseBuffApplyEvent(chain, event)
	default:
		return parseDirectDamageEvent(chain, event)
	}
}

type TrackingChangedEvent struct {
	BaseEvent
	Tracking bool
}
type StateChangedEvent struct {
	BaseEvent
	Downed   bool
	Defeated bool
}
type EnterCombatEvent struct {
	BaseEvent
	Subgroup int
}
type ExitCombatEvent struct {
	BaseEvent
}
type HealthUpdateEvent struct {
	BaseEvent
	// Percentage is fixed-point with two decimal places.
	// (99.5% is represented as 9950)
	Percentage uint16
}
type LogStartEvent struct {
	BaseEvent
}
type LogEndEvent struct {
	BaseEvent
	RealServerTime time.Time
	RealLocalTime  time.Time
}
type MaxHealthUpdateEvent struct {
	BaseEvent
	MaxHealth uint64
}
type RewardEvent struct {
	BaseEvent
	RewardID   int
	RewardType int
}
type PositionEvent struct {
	BaseEvent
	X, Y, Z float32
}
type VelocityEvent struct {
	BaseEvent
	X, Y, Z float32
}
type FacingEvent struct {
	BaseEvent
	X, Y float32
}
type TeamChangeEvent struct {
	BaseEvent
	TeamID int
}
type GuildEvent struct {
	BaseEvent
	Guild uuid.UUID
}
type SkillActivationEvent struct {
	CommonEvent
	ExpectedDuration time.Duration
	Quickness        bool
}
type SkillActivatedEvent struct {
	CommonEvent
	Duration time.Duration
	Complete bool
	Reset    bool
}
type BuffRemoveEvent struct {
	CommonEvent
	Duration    time.Duration
	Intensity   time.Duration
	Count       int
	Instance    uint32
	Synthesized bool
	All         bool
}
type ApplyBuffEvent struct {
	CommonEvent
	Duration time.Duration
	Instance uint32
	Active   bool

	WastedDuration time.Duration
	NewDuration    time.Duration
}
type BuffDamageEvent struct {
	CommonEvent
	Damage     int
	Tick       bool
	Success    bool
	invulnType uint8
}
type DirectDamageEvent struct {
	CommonEvent
	Damage         int
	Barrier        int
	WasDowned      bool
	Success        bool
	Critical       bool
	Glancing       bool
	Interrupt      bool
	Blocked        bool
	Evaded         bool
	Invulnerable   bool
	Missed         bool
	BecameDefeated bool
	BecameDowned   bool
}
type InitialBuffEvent struct {
	BaseEvent
	SkillID   int
	SkillName string
	Duration  time.Duration
	Instance  uint32
	Active    bool
}

func (e *InitialBuffEvent) Skill() (id int, name string) {
	return e.SkillID, e.SkillName
}

type WeaponSwapEvent struct {
	BaseEvent
	WeaponSet int
}
type BuffActiveEvent struct {
	BaseEvent
	Instance uint32
}
type BuffResetEvent struct {
	BaseEvent
	Duration time.Duration
	Instance uint32
}
type WeakPointEvent struct {
	BaseEvent
	Boss       *Agent
	Targetable bool
}
type TargetableEvent struct {
	BaseEvent
	Targetable bool
}

func parseStateChangeEvent(chain *EventChain, event cbtevent1) (Event, error) {
	switch event.IsStateChange {
	case 1: // CBTS_ENTERCOMBAT, src_agent entered combat, dst_agent is subgroup
		return &EnterCombatEvent{
			BaseEvent: makeBaseEvent("EnterCombat", chain, event),
			Subgroup:  int(event.DstAgent),
		}, nil
	case 2: // CBTS_EXITCOMBAT, src_agent left combat
		return &ExitCombatEvent{
			BaseEvent: makeBaseEvent("ExitCombat", chain, event),
		}, nil
	case 3: // CBTS_CHANGEUP, src_agent is now alive
		return &StateChangedEvent{
			BaseEvent: makeBaseEvent("StateChanged", chain, event),
			Downed:    false,
			Defeated:  false,
		}, nil
	case 4: // CBTS_CHANGEDEAD, src_agent is now dead
		return &StateChangedEvent{
			BaseEvent: makeBaseEvent("StateChanged", chain, event),
			Downed:    false,
			Defeated:  true,
		}, nil
	case 5: // CBTS_CHANGEDOWN, src_agent is now downed
		return &StateChangedEvent{
			BaseEvent: makeBaseEvent("StateChanged", chain, event),
			Downed:    true,
			Defeated:  false,
		}, nil
	case 6: // CBTS_SPAWN, src_agent is now in game tracking range (not in realtime api)
		return &TrackingChangedEvent{
			BaseEvent: makeBaseEvent("TrackingChanged", chain, event),
			Tracking:  true,
		}, nil
	case 7: // CBTS_DESPAWN, src_agent is no longer being tracked (not in realtime api)
		return &TrackingChangedEvent{
			BaseEvent: makeBaseEvent("TrackingChanged", chain, event),
			Tracking:  false,
		}, nil
	case 8: // CBTS_HEALTHUPDATE, src_agent has reached a health marker. dst_agent = percent * 10000 (eg. 99.5% will be 9950) (not in realtime api)

		return &HealthUpdateEvent{
			BaseEvent:  makeBaseEvent("HealthUpdate", chain, event),
			Percentage: uint16(event.DstAgent),
		}, nil
	case 9: // CBTS_LOGSTART, log start. value = server unix timestamp **uint32**. buff_dmg = local unix timestamp. src_agent = 0x637261 (arcdps id)
		chain.serverTime = time.Unix(int64(uint32(event.Value)), 0).UTC()
		chain.localTime = time.Unix(int64(uint32(event.BuffDmg)), 0).UTC()
		chain.timeOffset = -time.Duration(event.Time) * time.Millisecond
		return &LogStartEvent{
			BaseEvent: BaseEvent{
				Type:       "LogStart",
				ServerTime: chain.serverTime,
				LocalTime:  chain.localTime,
			},
		}, nil
	case 10: // CBTS_LOGEND, log end. value = server unix timestamp **uint32**. buff_dmg = local unix timestamp. src_agent = 0x637261 (arcdps id)
		be := makeBaseEvent("LogEnd", chain, event)
		be.Source = nil // just to be safe
		return &LogEndEvent{
			BaseEvent:      be,
			RealServerTime: time.Unix(int64(uint32(event.Value)), 0).UTC(),
			RealLocalTime:  time.Unix(int64(uint32(event.BuffDmg)), 0).UTC(),
		}, nil
	case 11: // CBTS_WEAPSWAP, src_agent swapped weapon set. dst_agent = current set id (0/1 water, 4/5 land)
		return &WeaponSwapEvent{
			BaseEvent: makeBaseEvent("WeaponSwap", chain, event),
			WeaponSet: int(event.DstAgent),
		}, nil
	case 12: // CBTS_MAXHEALTHUPDATE, src_agent has had it's maximum health changed. dst_agent = new max health (not in realtime api)
		return &MaxHealthUpdateEvent{
			BaseEvent: makeBaseEvent("MaxHealthUpdate", chain, event),
			MaxHealth: event.DstAgent,
		}, nil
	case 13: // CBTS_POINTOFVIEW, src_agent is agent of "recording" player
		chain.PointOfView = chain.agents[event.SrcAgent]
		return nil, nil
	case 14: // CBTS_LANGUAGE, src_agent is text language
		switch event.SrcAgent {
		case 0:
			chain.Language = language.English
		case 1:
			chain.Language = language.Korean
		case 2:
			chain.Language = language.French
		case 3:
			chain.Language = language.German
		case 4:
			chain.Language = language.Spanish
		case 5:
			chain.Language = language.Chinese
		default:
			return nil, errors.Errorf("evtc: unknown language ID %d", event.SrcAgent)
		}
		return nil, nil
	case 15: // CBTS_GWBUILD, src_agent is game build
		chain.BuildID = int(event.SrcAgent)
		return nil, nil
	case 16: // CBTS_SHARDID, src_agent is sever shard id
		chain.WorldID = uint16(event.SrcAgent)
		return nil, nil
	case 17: // CBTS_REWARD, src_agent is self, dst_agent is reward id, value is reward type. these are the wiggly boxes that you get
		return &RewardEvent{
			BaseEvent:  makeBaseEvent("Reward", chain, event),
			RewardID:   int(event.DstAgent),
			RewardType: int(event.Value),
		}, nil
	case 18: // CBTS_BUFFINITIAL, combat event that will appear once per buff per agent on logging start (statechange==18, buff==18, normal cbtevent otherwise)
		e, err := parseBuffApplyEvent(chain, event)
		if err != nil {
			return nil, err
		}

		abe := e.(*ApplyBuffEvent)
		return &InitialBuffEvent{
			BaseEvent: abe.BaseEvent,
			SkillID:   abe.SkillID,
			SkillName: abe.SkillName,
			Duration:  abe.Duration,
			Instance:  abe.Instance,
			Active:    abe.Active,
		}, nil
	case 19: // CBTS_POSITION, src_agent changed, cast float* p = (float*)&dst_agent, access as x/y/z (float[3]) (not in realtime api)
		return &PositionEvent{
			BaseEvent: makeBaseEvent("Position", chain, event),
			X:         math.Float32frombits(uint32(event.DstAgent)),
			Y:         math.Float32frombits(uint32(event.DstAgent >> 32)),
			Z:         math.Float32frombits(uint32(event.Value)),
		}, nil
	case 20: // CBTS_VELOCITY, src_agent changed, cast float* v = (float*)&dst_agent, access as x/y/z (float[3]) (not in realtime api)
		return &VelocityEvent{
			BaseEvent: makeBaseEvent("Velocity", chain, event),
			X:         math.Float32frombits(uint32(event.DstAgent)),
			Y:         math.Float32frombits(uint32(event.DstAgent >> 32)),
			Z:         math.Float32frombits(uint32(event.Value)),
		}, nil
	case 21: // CBTS_FACING, src_agent changed, cast float* f = (float*)&dst_agent, access as x/y (float[2]) (not in realtime api)
		return &FacingEvent{
			BaseEvent: makeBaseEvent("Facing", chain, event),
			X:         math.Float32frombits(uint32(event.DstAgent)),
			Y:         math.Float32frombits(uint32(event.DstAgent >> 32)),
		}, nil
	case 22: // CBTS_TEAMCHANGE, src_agent change, dst_agent new team id
		return &TeamChangeEvent{
			BaseEvent: makeBaseEvent("TeamChange", chain, event),
			TeamID:    int(event.DstAgent),
		}, nil
	case 23: // CBTS_ATTACKTARGET, src_agent is an attacktarget, dst_agent is the parent agent (gadget type), value is the current targetable state (not in realtime api)
		return &WeakPointEvent{
			BaseEvent:  makeBaseEvent("WeakPoint", chain, event),
			Boss:       chain.agents[event.DstAgent],
			Targetable: event.Value != 0,
		}, nil
	case 24: // CBTS_TARGETABLE, dst_agent is new target-able state (0 = no, 1 = yes. default yes) (not in realtime api)
		return &TargetableEvent{
			BaseEvent:  makeBaseEvent("Targetable", chain, event),
			Targetable: event.DstAgent != 0,
		}, nil
	case 25: // CBTS_MAPID, src_agent is map id
		chain.MapID = uint16(event.SrcAgent)
		return nil, nil
	case 27: // CBTS_STACKACTIVE, src_agent is agent with buff, dst_agent is the stackid marked active
		return &BuffActiveEvent{
			BaseEvent: makeBaseEvent("BuffActive", chain, event),
			Instance:  uint32(event.DstAgent),
		}, nil
	case 28: // CBTS_STACKRESET, src_agent is agent with buff, value is the duration to reset to (also marks inactive), pad61- is the stackid
		return &BuffResetEvent{
			BaseEvent: makeBaseEvent("BuffReset", chain, event),
			Duration:  time.Duration(uint32(event.Value)) * time.Millisecond,
			Instance:  event.Pad61_64,
		}, nil
	case 29: // CBTS_GUILD, src_agent is agent, dst_agent through buff_dmg is 16 byte guid (client form, needs minor rearrange for api form)
		var guid uuid.UUID
		binary.LittleEndian.PutUint64(guid[:], event.DstAgent)
		binary.LittleEndian.PutUint32(guid[8:], uint32(event.Value))
		binary.LittleEndian.PutUint32(guid[12:], uint32(event.BuffDmg))

		guid[0], guid[3] = guid[3], guid[0]
		guid[1], guid[2] = guid[2], guid[1]
		guid[4], guid[5] = guid[5], guid[4]
		guid[6], guid[7] = guid[7], guid[6]

		return &GuildEvent{
			BaseEvent: makeBaseEvent("Guild", chain, event),
			Guild:     guid,
		}, nil
	default:
		// TODO: generic format for unhandled cbtstatechange events
		spew.Dump(event)
		panic("TODO")
	}
}

func parseActivationEvent(chain *EventChain, event cbtevent1) (Event, error) {
	switch event.IsActivation {
	case 1: // ACTV_NORMAL, started skill activation without quickness
		return &SkillActivationEvent{
			CommonEvent:      makeCommonEvent("SkillActivation", chain, event),
			ExpectedDuration: time.Duration(uint32(event.Value)) * time.Millisecond,
			Quickness:        false,
		}, nil
	case 2: // ACTV_QUICKNESS, started skill activation with quickness
		return &SkillActivationEvent{
			CommonEvent:      makeCommonEvent("SkillActivation", chain, event),
			ExpectedDuration: time.Duration(uint32(event.Value)) * time.Millisecond,
			Quickness:        true,
		}, nil
	case 3: // ACTV_CANCEL_FIRE, stopped skill animation with reaching tooltip time
		return &SkillActivatedEvent{
			CommonEvent: makeCommonEvent("SkillActivated", chain, event),
			Duration:    time.Duration(uint32(event.Value)) * time.Millisecond,
			Complete:    true,
			Reset:       false,
		}, nil
	case 4: // ACTV_CANCEL_CANCEL, stopped skill activation without reaching tooltip time
		return &SkillActivatedEvent{
			CommonEvent: makeCommonEvent("SkillActivated", chain, event),
			Duration:    time.Duration(uint32(event.Value)) * time.Millisecond,
			Complete:    false,
			Reset:       false,
		}, nil
	case 5: // ACTV_RESET, animation completed fully
		return &SkillActivatedEvent{
			CommonEvent: makeCommonEvent("SkillActivated", chain, event),
			Duration:    time.Duration(uint32(event.Value)) * time.Millisecond,
			Complete:    true,
			Reset:       true,
		}, nil
	default:
		spew.Dump(event)
		panic("TODO")
	}
}

func parseBuffRemoveEvent(chain *EventChain, event cbtevent1) (Event, error) {
	e := &BuffRemoveEvent{
		CommonEvent: makeCommonEvent("BuffRemove", chain, event),
		Duration:    time.Duration(event.Value) * time.Millisecond,
		Intensity:   time.Duration(event.BuffDmg) * time.Millisecond,
		Count:       int(event.Result),
		Instance:    event.Pad61_64,
	}

	// src_agent loses the buff, so this makes more sense
	e.Source, e.Target = e.Target, e.Source

	switch event.IsBuffRemove {
	case 1: // CBTB_ALL, last/all stacks removed (sent by server)
		e.Synthesized = false
		e.All = true
	case 2: // CBTB_SINGLE, single stack removed (sent by server). will happen for each stack on cleanse
		e.Synthesized = false
		e.All = false
	case 3: // CBTB_MANUAL, single stack removed (auto by arc on ooc or all stack, ignore for strip/cleanse calc, use for in/out volume)
		e.Synthesized = true
		e.All = false
	default:
		spew.Dump(event)
		panic("TODO")
	}

	return e, nil
}

func parseBuffApplyEvent(chain *EventChain, event cbtevent1) (Event, error) {
	var wastedDuration, newDuration time.Duration
	if event.IsOffCycle == 0 {
		wastedDuration = time.Duration(event.OverstackValue) * time.Millisecond
	} else {
		newDuration = time.Duration(event.OverstackValue) * time.Millisecond
	}

	return &ApplyBuffEvent{
		CommonEvent: makeCommonEvent("ApplyBuff", chain, event),
		Duration:    time.Duration(event.Value) * time.Millisecond,
		Instance:    event.Pad61_64,
		Active:      event.IsShields != 0,

		WastedDuration: wastedDuration,
		NewDuration:    newDuration,
	}, nil
}

func parseBuffDamageEvent(chain *EventChain, event cbtevent1) (Event, error) {
	return &BuffDamageEvent{
		CommonEvent: makeCommonEvent("BuffDamage", chain, event),
		Damage:      int(event.BuffDmg),
		Tick:        event.IsOffCycle == 0,
		Success:     event.Result == 0,
		invulnType:  event.Result, // TODO
	}, nil
}

func parseDirectDamageEvent(chain *EventChain, event cbtevent1) (Event, error) {
	e := &DirectDamageEvent{
		CommonEvent: makeCommonEvent("DirectDamage", chain, event),
		Damage:      int(event.Value),
		Barrier:     int(event.OverstackValue),
		WasDowned:   event.IsOffCycle != 0,
	}

	switch event.Result {
	case 0: // CBTR_NORMAL, good physical hit
		e.Success = true
	case 1: // CBTR_CRIT, physical hit was crit
		e.Success = true
		e.Critical = true
	case 2: // CBTR_GLANCE, physical hit was glance
		e.Success = true
		e.Glancing = true
	case 3: // CBTR_BLOCK, physical hit was blocked eg. mesmer shield 4
		e.Blocked = true
	case 4: // CBTR_EVADE, physical hit was evaded, eg. dodge or mesmer sword 2
		e.Evaded = true
	case 5: // CBTR_INTERRUPT, physical hit interrupted something
		e.Success = true
		e.Interrupt = true
	case 6: // CBTR_ABSORB, physical hit was "invlun" or absorbed eg. guardian elite
		e.Invulnerable = true
	case 7: // CBTR_BLIND, physical hit missed
		e.Missed = true
	case 8: // CBTR_KILLINGBLOW, hit was killing hit
		e.BecameDefeated = true
	case 9: // CBTR_DOWNED, hit was downing hit
		e.BecameDowned = true
	default:
		spew.Dump(event)
		panic("TODO")
	}

	return e, nil
}
