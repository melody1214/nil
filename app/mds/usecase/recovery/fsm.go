package recovery

// For state transition.
// Recovery domain accepts any fail or membership changed notification from
// other domains. The rpc handler will add the notification into the noti queues.
// Then the fsm, who has the responsibility for handling each notifications will
// take noti from queues and handle it with proper state transitioning.
type fsm func() (next fsm)

func (h *handlers) fsmRun() {

}
