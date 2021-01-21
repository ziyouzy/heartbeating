package heartbeating

const (
	TCP byte    = 0
	UDP byte  = 1
	SERIAL byte = 2
	SNMP byte = 3
)

func Def2String(byte)string{
	switch{
	case 0:
		return "TCP"
	case 1：
		return "UDP"
	case 2:
		return "SERIAL"
	case 3:
		return "SNMP"
	default:
		return "未知类型"
	}
}

const (
	STOP_BEFORE_EXPIRE = true
	STOP_AFTER_EXPIRE = false
)
