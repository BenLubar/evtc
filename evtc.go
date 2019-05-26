// Package evtc implements a parser for arcdps event chains (combat logs).
package evtc

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

// Parse parses and EVTC file.
func Parse(r io.Reader) (*EventChain, error) {
	h, agents, skills, err := parseHeader(r)
	if err != nil {
		return nil, err
	}

	var events []cbtevent1

	switch h.Revision {
	case 0:
		for {
			var event cbtevent0
			if err := binary.Read(r, binary.LittleEndian, &event); err == io.EOF {
				break
			} else if err != nil {
				return nil, errors.Wrap(err, "evtc: error reading events")
			}
			events = append(events, convert0(event))
		}
	case 1:
		for {
			var event cbtevent1
			if err := binary.Read(r, binary.LittleEndian, &event); err == io.EOF {
				break
			} else if err != nil {
				return nil, errors.Wrap(err, "evtc: error reading events")
			}
			events = append(events, event)
		}
	}

	wrappedAgents := wrapAgents(agents, events)

	return makeEventChain(h, wrappedAgents, skills, events)
}
