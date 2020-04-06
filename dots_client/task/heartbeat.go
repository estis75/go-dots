package task

import (
    "time"
    "github.com/nttdots/go-dots/libcoap"
    "github.com/nttdots/go-dots/dots_common/messages"
    log "github.com/sirupsen/logrus"
)

type HeartBeatResponseHandler func(*HeartBeatTask, *libcoap.Pdu)
type HeartBeatTimeoutHandler  func(*HeartBeatTask, *Env)

type HeartBeatTask struct {
    TaskBase

    interval        time.Duration
    retry           int
    timeout         time.Duration
    responseHandler HeartBeatResponseHandler
    timeoutHandler  HeartBeatTimeoutHandler
    current_hb_id  string
}

type HeartBeatEvent struct { EventBase }

var isReceiveResponseContent bool
var isReceiveHeartBeat bool

func SetIsReceiveResponseContent(isContent bool) {
    isReceiveResponseContent = isContent
}

func SetIsReceiveHeartBeat(ishb bool) {
    isReceiveHeartBeat = ishb
}

func NewHeartBeatTask(interval time.Duration, retry int, timeout time.Duration, responseHandler HeartBeatResponseHandler, timeoutHandler HeartBeatTimeoutHandler) *HeartBeatTask {
    return &HeartBeatTask {
        newTaskBase(),
        interval,
        retry,
        timeout,
        responseHandler,
        timeoutHandler,
        "",
    }
}

func (t *HeartBeatTask) run(out chan Event) {
    for {
        select {
        case <- t.stopChan:{
            log.Debug("Current heartbeat task ended.")
            return
        }
        case <- time.After(t.interval):
            out <- &HeartBeatEvent{ EventBase{ t } }
        }
    }
}

func (e *HeartBeatEvent) Handle(env *Env) {
    hbTask := e.task.(*HeartBeatTask)
    currentTask := env.requests[hbTask.current_hb_id]

    if env.GetIsServerStopped() {
        log.Warn("Stopped heartbeat task.")
        env.StopHeartBeat()
        return
    }
    if currentTask != nil {
        log.Debugf("Waiting for current heartbeat message to be completed...")
        return
    }

    // If DOTS client receives 2.04 but DOTS client doesn't recieve heartbeat from DOTS server,  DOTS client set 'peer-hb-status' to false
    // Else  DOTS client set 'peer-hb-status' to true
    hbValue := true
    if isReceiveResponseContent && !isReceiveHeartBeat {
        hbValue  = false
    }

    // Send new heartbeat message
    pdu, err := messages.NewHeartBeatMessage(*env.CoapSession(), messages.JSON_HEART_BEAT_CLIENT, hbValue)
    if err != nil {
        log.Errorf("Failed to create heartbeat message")
        return
    }

    isReceiveHeartBeat = false
    isReceiveResponseContent =false
    task := e.Task().(*HeartBeatTask)
    newTask := NewMessageTask(
        pdu,
		task.interval,
		task.retry,
		task.timeout,
        false,
        true,
        func (_ *MessageTask, pdu *libcoap.Pdu, env *Env) {
            task.responseHandler(task, pdu)
        },
        func (*MessageTask, *Env) {
            task.timeoutHandler(task, env)
        })
    env.Run(newTask)
    hbTask.current_hb_id = newTask.message.AsMapKey()
	log.Debugf ("Sent new heartbeat message (id = %+v)", newTask.message.MessageID )
	log.Debugf ("pdu = %+v", pdu )
}

func (t * HeartBeatTask) IsRunnable() bool {
    return t.interval > 0
}