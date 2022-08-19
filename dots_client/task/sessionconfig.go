package task

import "time"
import "github.com/nttdots/go-dots/libcoap"
import log "github.com/sirupsen/logrus"
import "fmt"

type SessionConfigResponseHandler func(*SessionConfigTask, *libcoap.Pdu, *Env)
type SessionConfigTimeoutHandler  func(*SessionConfigTask, *Env)

type SessionConfigTask struct {
    TaskBase

    message         *libcoap.Pdu
    interval        time.Duration
    responseHandler SessionConfigResponseHandler
    timeoutHandler  SessionConfigTimeoutHandler
    current_sessionconfig_id  string
}

type SessionConfigEvent struct { EventBase }


func NewSessionConfigTask(message  *libcoap.Pdu, interval time.Duration, responseHandler SessionConfigResponseHandler, timeoutHandler SessionConfigTimeoutHandler) *SessionConfigTask {
	return &SessionConfigTask {
        newTaskBase(),
        message,
        interval,
        responseHandler,
        timeoutHandler,
        "0",  // fake value to avoid collision with id of ping task 
	}
}

func (t *SessionConfigTask) run(out chan Event) {
    if t.interval == 0 {
        out <- &SessionConfigEvent{ EventBase{ t } }
        return
    }
    for {
        select {
        case <- t.stopChan:{
            log.Debug("Current session config task ended.")
            return
        }
        case <- time.After(t.interval):
            log.Infof("Refresh session config after time = %+v", t.interval)
            out <- &SessionConfigEvent{ EventBase{ t } }
        }
    }
}

func (e *SessionConfigEvent) Handle(env *Env) {
    sessionConfigTask := e.task.(*SessionConfigTask)
    currentTask := env.requests[sessionConfigTask.current_sessionconfig_id]

    if currentTask != nil {
        log.Debugf("Waiting for current session config message (id=%+v)to be completed...", sessionConfigTask.current_sessionconfig_id)
        return
    }
    task := e.Task().(*SessionConfigTask)
    task.message.MessageID = env.session.NewMessageID()
    newTask := NewMessageTask(
        task.message,
        time.Duration(0),
        0,
        time.Duration(0),
        false,
        false,
        func (_ *MessageTask, pdu *libcoap.Pdu, env *Env) {
            task.responseHandler(task, pdu, env)
        },
        func (*MessageTask, *Env) {
            task.timeoutHandler(task, env)
        })
    env.Run(newTask)
    task.current_sessionconfig_id = fmt.Sprintf("%d", newTask.message.MessageID)
}

func (t *SessionConfigTask) MessageTask() *libcoap.Pdu {
    return t.message
}