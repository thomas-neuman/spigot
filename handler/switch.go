package handler

type LogicalPort struct {
	BasePacketHandler
}

type LogicalSwitch struct {
	BasePacketHandler
}

type LogicalRouter struct {
	BasePacketHandler
}

func (lp *LogicalPort) AttachSwitch(ls *LogicalSwitch) {
	lpp := lp.AddPort()
	_ = ls.AttachPort(lp, lpp.Id())
}

func (ls *LogicalSwitch) AttachPort(lp *LogicalPort, lpp PortId) PortId {
	inport := ls.AddPort()

	match := func(*PacketMetadata) bool {
		return true
	}
	ls.parser.AddRule(1, lpp, match)

	return inport.Id()
}

func (ls *LogicalSwitch) AttachRouter(lr *LogicalRouter) {
	lsp := ls.AddPort()
	_ = lr.AttachSwitch(ls, lsp.Id())
}

func (lr *LogicalRouter) AttachSwitch(ls *LogicalSwitch, lsp PortId) PortId {
	inport := lr.AddPort()

	match := func(*PacketMetadata) bool {
		return true
	}
	lr.parser.AddRule(1, lsp, match)

	return inport.Id()
}
