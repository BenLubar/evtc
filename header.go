package evtc

import (
	"encoding/binary"
	"io"
	"strings"

	"github.com/pkg/errors"
)

type header struct {
	Magic    [4]byte // {'E', 'V', 'T', 'C'}
	Date     [8]byte // arcdps build datestamp
	Revision uint8   // 0 or 1; decides which cbtevent struct is used
	Boss     uint16  // species ID
	Reserved uint8   // unused; reserved
}

func parseHeader(r io.Reader) (header, []agent, map[uint32]string, error) {
	var h header
	if err := errors.Wrap(binary.Read(r, binary.LittleEndian, &h), "evtc: could not read header"); err != nil {
		return header{}, nil, nil, err
	}
	if h.Magic[0] != 'E' || h.Magic[1] != 'V' || h.Magic[2] != 'T' || h.Magic[3] != 'C' {
		return header{}, nil, nil, errors.Errorf("evtc: invalid magic number (expecting \"EVTC\"): %q", h.Magic[:])
	}
	var count uint32
	if err := errors.Wrap(binary.Read(r, binary.LittleEndian, &count), "evtc: could not read agent count"); err != nil {
		return header{}, nil, nil, err
	}
	agents := make([]agent, count)
	if err := errors.Wrap(binary.Read(r, binary.LittleEndian, agents), "evtc: could not read agents"); err != nil {
		return header{}, nil, nil, err
	}
	if err := errors.Wrap(binary.Read(r, binary.LittleEndian, &count), "evtc: could not read skill count"); err != nil {
		return header{}, nil, nil, err
	}
	skills := make([]skill, count)
	if err := errors.Wrap(binary.Read(r, binary.LittleEndian, skills), "evtc: could not read skills"); err != nil {
		return header{}, nil, nil, err
	}

	return h, agents, wrapSkills(skills), nil
}

type skill struct {
	ID   uint32
	Name [64]byte
}

func wrapSkills(skills []skill) map[uint32]string {
	m := map[uint32]string{
		1066:  "Resurrect", // not custom but important and unnamed
		1175:  "Bandage",   // personal healing only
		65001: "Dodge",     // will occur in is_activation==normal event
	}

	for _, s := range skills {
		name := strings.SplitN(string(s.Name[:]), "\x00", 2)
		m[s.ID] = name[0]
	}

	return m
}
