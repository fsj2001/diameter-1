package sock

import (
	"time"

	"github.com/fkgi/diameter/msg"
	"github.com/fkgi/diameter/rfc6733"
)

// UnknownIDAnswer is error
type UnknownIDAnswer struct {
	msg.RawMsg
}

func (e UnknownIDAnswer) Error() string {
	return "Unknown message recieved"
}

// FailureAnswer is error
type FailureAnswer struct {
	msg.RawMsg
}

func (e FailureAnswer) Error() string {
	return "Answer message with failure"
}

// RcvCER
type eventRcvCER struct {
	m msg.RawMsg
}

func (eventRcvCER) String() string {
	return "Rcv-CER"
}

func (v eventRcvCER) exec(c *Conn) error {
	if c.state != waitCER {
		return NotAcceptableEvent{stateEvent: v, state: c.state}
	}

	cer, e := rfc6733.CER{}.FromRaw(v.m)
	Notify(CapabilityExchangeEvent{tx: false, req: true, conn: c, Err: e})

	if e != nil {
		// ToDo
		// make error answer for undecodable CER
		c.con.Close()
		return e
	}

	cea := HandleCER(cer.(rfc6733.CER), c)
	m := cea.ToRaw()
	m.HbHID = v.m.HbHID
	m.EtEID = v.m.EtEID
	if cea.ResultCode != rfc6733.DiameterSuccess {
		m.FlgE = true
	}
	c.con.SetWriteDeadline(time.Now().Add(TransportTimeout))
	_, e = m.WriteTo(c.con)

	if e == nil && cea.ResultCode != rfc6733.DiameterSuccess {
		e = FailureAnswer{m}
	}
	if e == nil {
		c.state = open
		c.sysTimer = time.AfterFunc(c.peer.WDInterval, func() {
			c.watchdog()
		})
	}

	Notify(CapabilityExchangeEvent{tx: true, req: false, conn: c, Err: e})
	if e != nil {
		c.con.Close()
	}
	return e
}

// RcvCEA
type eventRcvCEA struct {
	m msg.RawMsg
}

func (eventRcvCEA) String() string {
	return "Rcv-CEA"
}

func (v eventRcvCEA) exec(c *Conn) error {
	if c.state != waitCEA {
		return NotAcceptableEvent{stateEvent: v, state: c.state}
	}
	ch, ok := c.sndstack[v.m.HbHID]
	if !ok {
		return UnknownIDAnswer{v.m}
	}
	delete(c.sndstack, v.m.HbHID)

	cea, e := rfc6733.CEA{}.FromRaw(v.m)
	if e == nil {
		HandleCEA(cea.(rfc6733.CEA), c)
		if cea.Result() == uint32(rfc6733.DiameterSuccess) {
			c.state = open
			c.sysTimer = time.AfterFunc(c.peer.WDInterval, func() {
				c.watchdog()
			})
		} else {
			e = FailureAnswer{v.m}
		}
	}

	Notify(CapabilityExchangeEvent{tx: false, req: false, conn: c, Err: e})
	if e != nil {
		c.con.Close()
		v.m = msg.RawMsg{}
	}
	ch <- v.m
	return e
}

type eventRcvDWR struct {
	m msg.RawMsg
}

func (eventRcvDWR) String() string {
	return "Rcv-DWR"
}

func (v eventRcvDWR) exec(c *Conn) error {
	if c.state != open {
		return NotAcceptableEvent{stateEvent: v, state: c.state}
	}

	dwr, e := rfc6733.DWR{}.FromRaw(v.m)
	Notify(WatchdogEvent{tx: false, req: true, conn: c, Err: e})

	if e != nil {
		// ToDo
		// make error answer for undecodable CER
		return e
	}

	dwa := HandleDWR(dwr.(rfc6733.DWR), c)
	m := dwa.ToRaw()
	m.HbHID = v.m.HbHID
	m.EtEID = v.m.EtEID
	if dwa.ResultCode != rfc6733.DiameterSuccess {
		m.FlgE = true
	}
	c.con.SetWriteDeadline(time.Now().Add(TransportTimeout))
	_, e = m.WriteTo(c.con)

	if e == nil && dwa.ResultCode != rfc6733.DiameterSuccess {
		e = FailureAnswer{m}
	}
	if e == nil {
		c.sysTimer.Reset(c.peer.WDInterval)
	}

	Notify(WatchdogEvent{tx: true, req: false, conn: c, Err: e})
	if e != nil {
		c.con.Close()
	}
	return e
}

// RcvDWA
type eventRcvDWA struct {
	m msg.RawMsg
}

func (eventRcvDWA) String() string {
	return "Rcv-DWA"
}

func (v eventRcvDWA) exec(c *Conn) error {
	if c.state != open {
		return NotAcceptableEvent{stateEvent: v, state: c.state}
	}
	ch, ok := c.sndstack[v.m.HbHID]
	if !ok {
		return UnknownIDAnswer{v.m}
	}
	delete(c.sndstack, v.m.HbHID)

	dwa, e := rfc6733.DWA{}.FromRaw(v.m)
	if e == nil {
		HandleDWA(dwa.(rfc6733.DWA), c)
		if dwa.Result() == uint32(rfc6733.DiameterSuccess) {
			c.wdCounter = 0
			c.sysTimer.Reset(c.peer.WDInterval)
		} else {
			e = FailureAnswer{v.m}
		}
	}

	Notify(WatchdogEvent{tx: false, req: false, conn: c, Err: e})
	if e != nil {
		v.m = msg.RawMsg{}
	}
	ch <- v.m
	return e
}

type eventRcvDPR struct {
	m msg.RawMsg
}

func (eventRcvDPR) String() string {
	return "Rcv-DPR"
}

func (v eventRcvDPR) exec(c *Conn) error {
	if c.state != open {
		return NotAcceptableEvent{stateEvent: v, state: c.state}
	}

	dpr, e := rfc6733.DPR{}.FromRaw(v.m)
	Notify(PurgeEvent{tx: false, req: true, conn: c, Err: e})

	if e != nil {
		// ToDo
		// make error answer for undecodable CER
		return e
	}

	dpa := HandleDPR(dpr.(rfc6733.DPR), c)
	m := dpa.ToRaw()
	m.HbHID = v.m.HbHID
	m.EtEID = v.m.EtEID
	if dpa.ResultCode != rfc6733.DiameterSuccess {
		m.FlgE = true
	} else {
		c.state = closing
		c.sysTimer.Stop()
		c.sysTimer = time.AfterFunc(c.peer.SndTimeout, func() {
			c.con.Close()
		})
	}
	c.con.SetWriteDeadline(time.Now().Add(TransportTimeout))
	_, e = m.WriteTo(c.con)

	Notify(&PurgeEvent{tx: true, req: false, conn: c, Err: e})
	if e != nil {
		c.con.Close()
	}
	return e
}

type eventRcvDPA struct {
	m msg.RawMsg
}

func (eventRcvDPA) String() string {
	return "Rcv-DPA"
}

func (v eventRcvDPA) exec(c *Conn) error {
	if c.state != closing {
		return NotAcceptableEvent{stateEvent: v, state: c.state}
	}
	ch, ok := c.sndstack[v.m.HbHID]
	if !ok {
		return UnknownIDAnswer{v.m}
	}
	delete(c.sndstack, v.m.HbHID)

	dpa, e := rfc6733.DPA{}.FromRaw(v.m)
	if e == nil {
		HandleDPA(dpa.(rfc6733.DPA), c)
		if dpa.Result() != uint32(rfc6733.DiameterSuccess) {
			e = FailureAnswer{v.m}
		}
	}

	Notify(PurgeEvent{tx: false, req: false, conn: c, Err: e})
	c.con.Close()
	if e != nil {
		v.m = msg.RawMsg{}
	}
	ch <- v.m
	return e
}

type eventRcvMsg struct {
	m msg.RawMsg
}

func (eventRcvMsg) String() string {
	return "Rcv-MSG"
}

func (v eventRcvMsg) exec(c *Conn) (e error) {
	if c.state != open {
		return NotAcceptableEvent{stateEvent: v, state: c.state}
	}

	if v.m.FlgR {
		c.rcvstack <- v.m
	} else {
		ch, ok := c.sndstack[v.m.HbHID]
		if !ok {
			return
		}
		delete(c.sndstack, v.m.HbHID)
		ch <- v.m
	}
	c.sysTimer.Reset(c.peer.WDInterval)

	Notify(MessageEvent{tx: false, req: true, conn: c, Err: e})
	return
}
