package diameter

import (
	"bytes"
	"fmt"
	"log"
)

// Notify is called when error or trace event are occured
var Notify = func(n Notice) {
	log.Println(n)
}

// Notice is notification information from connection
type Notice interface {
	fmt.Stringer
}

// StateUpdate notify event
type StateUpdate struct {
	oldStat state
	newStat state
	stateEvent
	conn *Conn
	Err  error
}

func (e StateUpdate) String() string {
	w := new(bytes.Buffer)
	fmt.Fprintf(w, "Event %s: Peer %s", e.stateEvent, e.conn.Peer)
	if e.oldStat != e.newStat {
		fmt.Fprintf(w, ": State %s -> %s", e.oldStat, e.newStat)
	} else {
		fmt.Fprintf(w, ": State %s", e.oldStat)
	}
	if e.Err != nil {
		fmt.Fprintf(w, ": Failed: %s", e.Err)
	}
	return w.String()
}

func msgHandleLog(x, r bool, c *Conn, e error, req, ans string) string {
	w := new(bytes.Buffer)
	if x {
		fmt.Fprintf(w, "-> ")
	} else {
		fmt.Fprintf(w, "<- ")
	}
	if r {
		fmt.Fprintf(w, req)
	} else {
		fmt.Fprintf(w, ans)
	}
	fmt.Fprintf(w, " (%s)", c.Peer)
	if e != nil {
		fmt.Fprintf(w, ": Failed: %s", e)
	}
	return w.String()
}

// CapabilityExchangeEvent notify capability exchange related event
type CapabilityExchangeEvent struct {
	tx   bool
	req  bool
	conn *Conn
	Err  error
}

func (e CapabilityExchangeEvent) String() string {
	return msgHandleLog(e.tx, e.req, e.conn, e.Err, "CER", "CEA")
}

// WatchdogEvent notify watchdog related event
type WatchdogEvent struct {
	tx   bool
	req  bool
	conn *Conn
	Err  error
}

func (e WatchdogEvent) String() string {
	return msgHandleLog(e.tx, e.req, e.conn, e.Err, "DWR", "DWA")
}

// MessageEvent notify diameter message related event
type MessageEvent struct {
	tx   bool
	req  bool
	conn *Conn
	Err  error
}

func (e MessageEvent) String() string {
	return msgHandleLog(e.tx, e.req, e.conn, e.Err, "REQ", "ANS")
}

// PurgeEvent notify diameter purge related event
type PurgeEvent struct {
	tx   bool
	req  bool
	conn *Conn
	Err  error
}

func (e PurgeEvent) String() string {
	return msgHandleLog(e.tx, e.req, e.conn, e.Err, "DPR", "DPA")
}
