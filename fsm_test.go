package ssfsm

import (
	"strings"
	"sync"
	"testing"
	"time"

	"fmt"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestSyncFSM(t *testing.T) {
	var lgf = func(fsm *FSM, tx Transition) error { return nil }
	fsm := NewFSM(false, "START", Table{
		Transition{"INCOMMING", "START", "WAITING"}:   lgf,
		Transition{"RECEIVING", "WAITING", "WAITING"}: lgf,
		Transition{"RECEIVED", "WAITING", "IDLE"}:     lgf,
		Transition{"TIMEOUT", "IDLE", "FINAL"}:        lgf,
	})

	assert.Equal(t, "START", fsm.state)
	err := fsm.Trigger("RECEIVING")
	assert.Error(t, err)
	assert.Equal(t, ErrStateConflict, errors.Cause(err))
	assert.True(t, strings.Contains(err.Error(), "event: RECEIVING"))
	assert.True(t, strings.Contains(err.Error(), "state: START"))

	err = fsm.Trigger("NONSENSE")
	assert.Error(t, err)
	assert.Equal(t, ErrEventNotFound, errors.Cause(err))
	assert.True(t, strings.Contains(err.Error(), "event: NONSENSE"))

	err = fsm.Trigger("INCOMMING")
	assert.Nil(t, err)
	assert.Equal(t, "WAITING", fsm.state)

	err = fsm.Trigger("TIMEOUT")
	assert.Error(t, err)
	assert.Equal(t, ErrStateConflict, errors.Cause(err))
	assert.True(t, strings.Contains(err.Error(), "event: TIMEOUT"))
	assert.True(t, strings.Contains(err.Error(), "state: WAITING"))

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
	assert.Equal(t, ErrStateConflict, errors.Cause(err))
	assert.True(t, strings.Contains(err.Error(), "event: TIMEOUT"))
	assert.True(t, strings.Contains(err.Error(), "state: FINAL"))
}

func TestTransitionCallbackErrorSyncFSM(t *testing.T) {
	condition := false
	var lgf = func(fsm *FSM, tx Transition) error {
		if tx.Event == "COMPLETED" && !condition {
			return fmt.Errorf("NOT COMPLETE")
		}
		return nil
	}
	fsm := NewFSM(false, "START", Table{
		Transition{"INCOMMING", "START", "WAITING"}:   lgf,
		Transition{"RECEIVING", "WAITING", "WAITING"}: lgf,
		Transition{"COMPLETED", "WAITING", "IDLE"}:    lgf,
		Transition{"RECEIVED", "WAITING", "IDLE"}:     lgf,
		Transition{"RECEIVING", "IDLE", "WAITING"}:    lgf,
		Transition{"TIMEOUT", "IDLE", "FINAL"}:        lgf,
	})

	assert.Equal(t, "START", fsm.state)
	err := fsm.Trigger("RECEIVING")
	assert.Error(t, err)
	assert.Equal(t, ErrStateConflict, errors.Cause(err))
	assert.True(t, strings.Contains(err.Error(), "event: RECEIVING"))
	assert.True(t, strings.Contains(err.Error(), "state: START"))

	err = fsm.Trigger("NONSENSE")
	assert.Error(t, err)
	assert.Equal(t, ErrEventNotFound, errors.Cause(err))
	assert.True(t, strings.Contains(err.Error(), "event: NONSENSE"))

	err = fsm.Trigger("INCOMMING")
	assert.Nil(t, err)
	assert.Equal(t, "WAITING", fsm.state)

	err = fsm.Trigger("TIMEOUT")
	assert.Error(t, err)
	assert.Equal(t, ErrStateConflict, errors.Cause(err))
	assert.True(t, strings.Contains(err.Error(), "event: TIMEOUT"))
	assert.True(t, strings.Contains(err.Error(), "state: WAITING"))

	err = fsm.Trigger("RECEIVING")
	assert.Nil(t, err)
	assert.Equal(t, "WAITING", fsm.state)

	err = fsm.Trigger("RECEIVING")
	assert.Nil(t, err)
	assert.Equal(t, "WAITING", fsm.state)

	err = fsm.Trigger("COMPLETED")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "NOT COMPLETE"))
	assert.Equal(t, "WAITING", fsm.state)

	condition = true
	err = fsm.Trigger("COMPLETED")
	assert.Nil(t, err)
	assert.Equal(t, "IDLE", fsm.state)

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
	assert.Equal(t, ErrStateConflict, errors.Cause(err))
	assert.True(t, strings.Contains(err.Error(), "event: TIMEOUT"))
	assert.True(t, strings.Contains(err.Error(), "state: FINAL"))
}

func TestAsyncFSM(t *testing.T) {
	var lgf = func(fsm *FSM, tx Transition) error { return nil }
	fsm := NewFSM(true, "START", Table{
		Transition{"INCOMMING", "START", "WAITING"}:   lgf,
		Transition{"RECEIVING", "WAITING", "WAITING"}: lgf,
		Transition{"RECEIVED", "WAITING", "IDLE"}:     lgf,
		Transition{"TIMEOUT", "IDLE", "FINAL"}:        lgf,
	})

	assert.Equal(t, "START", fsm.state)
	err := fsm.Trigger("RECEIVING")
	assert.Error(t, err)
	assert.Equal(t, ErrStateConflict, errors.Cause(err))
	assert.True(t, strings.Contains(err.Error(), "event: RECEIVING"))
	assert.True(t, strings.Contains(err.Error(), "state: START"))

	err = fsm.Trigger("INCOMMING")
	assert.Nil(t, err)
	assert.Equal(t, "WAITING", fsm.state)

	err = fsm.Trigger("TIMEOUT")
	assert.Error(t, err)
	assert.Equal(t, ErrStateConflict, errors.Cause(err))
	assert.True(t, strings.Contains(err.Error(), "event: TIMEOUT"))
	assert.True(t, strings.Contains(err.Error(), "state: WAITING"))

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
	assert.Equal(t, ErrStateConflict, errors.Cause(err))
	assert.True(t, strings.Contains(err.Error(), "event: TIMEOUT"))
	assert.True(t, strings.Contains(err.Error(), "state: FINAL"))
}

func TestAsyncFSMWithCallback(t *testing.T) {
	cnt := 0
	mx := new(sync.Mutex)

	var lgf = func(fsm *FSM, tx Transition) error {
		if tx.Event == "RECEIVING" && cnt < 3 {
			mx.Lock()
			defer mx.Unlock()
			cnt++
			go func() {
				err := fsm.Trigger("RECEIVING")
				assert.Nil(t, err)
			}()

			err := fsm.Trigger("RECEIVING")
			assert.Error(t, err)
			assert.Equal(t, ErrTransitionConflict, errors.Cause(err))
		}
		return nil
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
	assert.Equal(t, ErrStateConflict, errors.Cause(err))
	assert.True(t, strings.Contains(err.Error(), "event: RECEIVING"))
	assert.True(t, strings.Contains(err.Error(), "state: START"))

	err = fsm.Trigger("INCOMMING")
	assert.Nil(t, err)
	assert.Equal(t, "WAITING", fsm.state)

	err = fsm.Trigger("TIMEOUT")
	assert.Error(t, err)
	assert.Equal(t, ErrStateConflict, errors.Cause(err))
	assert.True(t, strings.Contains(err.Error(), "event: TIMEOUT"))
	assert.True(t, strings.Contains(err.Error(), "state: WAITING"))

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
	assert.Equal(t, ErrStateConflict, errors.Cause(err))
	assert.True(t, strings.Contains(err.Error(), "event: TIMEOUT"))
	assert.True(t, strings.Contains(err.Error(), "state: FINAL"))

	assert.Equal(t, 3, cnt)
}
