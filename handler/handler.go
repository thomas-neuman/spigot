package handler

import (
	"context"
)

type PacketHandler interface {
	In() chan *PacketMetadata
	Process(*PacketMetadata) error
}

type BasePacketHandler struct {
	engine *PacketProcessingEngine
	in     chan *PacketMetadata
	parser PacketParser
}

func (h *BasePacketHandler) In() chan *PacketMetadata {
	return h.in
}

func (h *BasePacketHandler) Process(md *PacketMetadata) error {
	outport, err := h.parser.Parse(md)
	if err == nil {
		return h.engine.ports[outport].Submit(md)
	}
	return err
}

func (h *BasePacketHandler) AddPort() Port {
	return h.engine.NewPort(h)
}

func (h *BasePacketHandler) Start(ctxt context.Context) {
	go func() {
		var md *PacketMetadata
		for {
			select {
			case md = <-h.In():
				h.Process(md)
			case <-ctxt.Done():
				return
			}
		}
	}()
}
