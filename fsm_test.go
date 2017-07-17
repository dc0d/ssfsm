package ssfsm

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSyncFSM(t *testing.T) {
	var lgf = func(fsm *FSM, tx Transition) {}
	fsm := NewFSM(false, "START", Table{
		Transition{"INCOMMING", "START", "WAITING"}:   lgf,
		Transition{"RECEIVING", "WAITING", "WAITING"}: lgf,
		Transition{"RECEIVED", "WAITING", "IDLE"}:     lgf,
		Transition{"TIMEOUT", "IDLE", "FINAL"}:        lgf,
	})

	assert.Equal(t, "START", fsm.state)
	err := fsm.Trigger("RECEIVING")
	assert.Error(t, err)
	assert.Equal(t, "no state found for event: RECEIVING current state: START", err.Error())

	err = fsm.Trigger("INCOMMING")
	assert.Nil(t, err)
	assert.Equal(t, "WAITING", fsm.state)

	err = fsm.Trigger("TIMEOUT")
	assert.Error(t, err)
	assert.Equal(t, "no state found for event: TIMEOUT current state: WAITING", err.Error())

	err = fsm.Trigger("RECEIVING")
	assert.Nil(t, err)
	assert.Equal(t, "WAITING", fsm.state)

	err = fsm.Trigger("RECEIVING")
	assert.Nil(t, err)
	assert.Equal(t, "WAITING", fsm.state)

	err = fsm.Trigger("RECEIVED")
	assert.Nil(t, err)
	assert.Equal(t, "IDLE", fsm.state)

	err = fsm.Trigger("TIMEOUT")
	assert.Nil(t, err)
	assert.Equal(t, "FINAL", fsm.state)

	err = fsm.Trigger("TIMEOUT")
	assert.Error(t, err)
	assert.Equal(t, "no state found for event: TIMEOUT current state: FINAL", err.Error())
}

func TestAsyncFSM(t *testing.T) {
	var lgf = func(fsm *FSM, tx Transition) {}
	fsm := NewFSM(true, "START", Table{
		Transition{"INCOMMING", "START", "WAITING"}:   lgf,
		Transition{"RECEIVING", "WAITING", "WAITING"}: lgf,
		Transition{"RECEIVED", "WAITING", "IDLE"}:     lgf,
		Transition{"TIMEOUT", "IDLE", "FINAL"}:        lgf,
	})

	assert.Equal(t, "START", fsm.state)
	err := fsm.Trigger("RECEIVING")
	assert.Error(t, err)
	assert.Equal(t, "no state found for event: RECEIVING current state: START", err.Error())

	err = fsm.Trigger("INCOMMING")
	assert.Nil(t, err)
	assert.Equal(t, "WAITING", fsm.state)

	err = fsm.Trigger("TIMEOUT")
	assert.Error(t, err)
	assert.Equal(t, "no state found for event: TIMEOUT current state: WAITING", err.Error())

	err = fsm.Trigger("RECEIVING")
	assert.Nil(t, err)
	assert.Equal(t, "WAITING", fsm.state)

	err = fsm.Trigger("RECEIVING")
	assert.Nil(t, err)
	assert.Equal(t, "WAITING", fsm.state)

	err = fsm.Trigger("RECEIVED")
	assert.Nil(t, err)
	assert.Equal(t, "IDLE", fsm.state)

	err = fsm.Trigger("TIMEOUT")
	assert.Nil(t, err)
	assert.Equal(t, "FINAL", fsm.state)

	err = fsm.Trigger("TIMEOUT")
	assert.Error(t, err)
	assert.Equal(t, "no state found for event: TIMEOUT current state: FINAL", err.Error())
}

func TestAsyncFSMWithCallback(t *testing.T) {
	cnt := 0
	mx := new(sync.Mutex)

	var lgf = func(fsm *FSM, tx Transition) {
		if tx.Event == "RECEIVING" && cnt < 3 {
			mx.Lock()
			defer mx.Unlock()
			cnt++
			go func() {
				err := fsm.Trigger("RECEIVING")
				assert.Nil(t, err)
			}()
		}
	}

	fsm := NewFSM(true, "START", Table{
		Transition{"INCOMMING", "START", "WAITING"}:   lgf,
		Transition{"RECEIVING", "WAITING", "WAITING"}: lgf,
		Transition{"RECEIVED", "WAITING", "IDLE"}:     lgf,
		Transition{"TIMEOUT", "IDLE", "FINAL"}:        lgf,
	})

	assert.Equal(t, "START", fsm.state)
	err := fsm.Trigger("RECEIVING")
	assert.Error(t, err)
	assert.Equal(t, "no state found for event: RECEIVING current state: START", err.Error())

	err = fsm.Trigger("INCOMMING")
	assert.Nil(t, err)
	assert.Equal(t, "WAITING", fsm.state)

	err = fsm.Trigger("TIMEOUT")
	assert.Error(t, err)
	assert.Equal(t, "no state found for event: TIMEOUT current state: WAITING", err.Error())

	err = fsm.Trigger("RECEIVING")
	assert.Nil(t, err)
	assert.Equal(t, "WAITING", fsm.state)

	<-time.After(time.Millisecond * 100)

	err = fsm.Trigger("RECEIVING")
	assert.Nil(t, err)
	assert.Equal(t, "WAITING", fsm.state)

	err = fsm.Trigger("RECEIVED")
	assert.Nil(t, err)
	assert.Equal(t, "IDLE", fsm.state)

	err = fsm.Trigger("TIMEOUT")
	assert.Nil(t, err)
	assert.Equal(t, "FINAL", fsm.state)

	err = fsm.Trigger("TIMEOUT")
	assert.Error(t, err)
	assert.Equal(t, "no state found for event: TIMEOUT current state: FINAL", err.Error())

	assert.Equal(t, 3, cnt)
}
