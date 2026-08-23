package heartbeat

import "dhs/internal/model"

const (
	ReasonRegister       = "register"
	ReasonFirstHeartbeat = "first_heartbeat"
	ReasonTimeout        = "heartbeat_timeout"
	ReasonRecoveryStart  = "recovery_start"
	ReasonRecovered      = "heartbeat_recovered"
	ReasonRetired        = "retired"
)

func IsOperational(s model.Status) bool {
	return s == model.Registered || s == model.Online || s == model.Recovering
}
func IsRecoverable(s model.Status) bool { return s == model.Lost || s == model.Recovering }
func IsTerminal(s model.Status) bool    { return s == model.Offline }
