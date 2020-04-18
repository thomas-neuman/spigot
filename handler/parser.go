package handler

type Priority uint8
type MatchCriteria func(*PacketMetadata) bool

type PacketParser interface {
	AddRule(prio Priority, outport PortId, match ...MatchCriteria) error
	RemoveRule(prio Priority, outport PortId, match ...MatchCriteria) error
	Parse(*PacketMetadata) (PortId, error)
}

type Rule struct{}

type BasePacketParser struct {
	rules []Rule
}

func (p *BasePacketParser) Parse(md *PacketMetadata) (PortId, error) {
	return PortId(0), nil
}

func (p *BasePacketParser) AddRule(prio Priority, outport PortId, match ...MatchCriteria) error {
	return nil
}

func (p *BasePacketParser) RemoveRule(prio Priority, outport PortId, match ...MatchCriteria) error {
	return nil
}
