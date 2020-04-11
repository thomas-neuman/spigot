package handler

import (
	"context"
	"errors"
	"log"

	"github.com/google/gopacket"
)

type LayerHandlerImpl interface {
	FirstLayerType() gopacket.LayerType
	DoInit() error
	DoRead() ([]gopacket.SerializableLayer, error)
	DoWrite([]gopacket.SerializableLayer) error
}

type LayerHandler interface {
	FirstLayerType() gopacket.LayerType
	Read() ([]gopacket.SerializableLayer, error)
	Write([]gopacket.SerializableLayer) error
	Start() error
}

type BaseLayerHandler struct {
	inbox  chan []gopacket.SerializableLayer
	outbox chan []gopacket.SerializableLayer
	ctxt   context.Context

	impl LayerHandlerImpl
}

func (h *BaseLayerHandler) Inbox() chan<- []gopacket.SerializableLayer {
	return h.inbox
}

func (h *BaseLayerHandler) Outbox() chan<- []gopacket.SerializableLayer {
	return h.outbox
}

func (h *BaseLayerHandler) FirstLayerType() gopacket.LayerType {
	return h.impl.FirstLayerType()
}

func (h *BaseLayerHandler) Read() (data []gopacket.SerializableLayer, err error) {
	select {
	case data = <-h.inbox:
		err = nil
		return
	case <-h.ctxt.Done():
		err = errors.New("Closed")
		return
	}
}

func (h *BaseLayerHandler) Write(data []gopacket.SerializableLayer) error {
	if len(data) < 1 {
		return errors.New("Nil layer slice")
	}

	if data[0].LayerType() == h.impl.FirstLayerType() {
		h.outbox <- data
		return nil
	}

	return errors.New("Cannot accept layer slice")
}

func (h *BaseLayerHandler) Start() error {
	err := h.impl.DoInit()
	if err != nil {
		return err
	}

	go func() {
		var msg []gopacket.SerializableLayer
		var err error

		for {
			select {
			case msg = <-h.outbox:
				err = h.impl.DoWrite(msg)
				if err != nil {
					log.Println("Error writing packet:", err)
				}
			case <-h.ctxt.Done():
				return
			}
		}
	}()

	go func() {
		for {
			var msg []gopacket.SerializableLayer
			var err error

			select {
			case <-h.ctxt.Done():
				return
			default:
				msg, err = h.impl.DoRead()
				if err != nil {
					log.Println("Error reading packet:", err)
					continue
				}
				h.inbox <- msg
			}
		}
	}()

	return nil
}

func NewLayerHandler(impl LayerHandlerImpl, ctxt context.Context) LayerHandler {
	return &BaseLayerHandler{
		inbox:  make(chan []gopacket.SerializableLayer),
		outbox: make(chan []gopacket.SerializableLayer),
		ctxt:   ctxt,
		impl:   impl,
	}
}
