package handler

type PortId uint16

type Port interface {
	Id() PortId
	Submit(*PacketMetadata) error
}

type BasePort struct {
	id      PortId
	handler PacketHandler
	packets chan *PacketMetadata
}

func (p *BasePort) Id() PortId {
	return p.id
}

func (p *BasePort) Submit(md *PacketMetadata) error {
	go func() {
		md.InPort = p.Id()
		p.handler.In() <- md
	}()

	return nil
}
