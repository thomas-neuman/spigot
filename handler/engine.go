package handler

type PacketProcessingEngine struct {
	ports []Port
}

func (e *PacketProcessingEngine) NewPort(handler PacketHandler) Port {
	pid := len(e.ports)
	inport := &BasePort{
		handler: handler,
		id:      PortId(pid),
	}
	e.ports = append(e.ports, inport)

	return inport
}

func (e *PacketProcessingEngine) NewPacketHandler(parser PacketParser) PacketHandler {
	handler := &BasePacketHandler{
		engine: e,
		in:     make(chan *PacketMetadata),
		parser: parser,
	}

	go func() {
		for {
		}
	}()

	return handler
}

func (e *PacketProcessingEngine) Start() {
}
