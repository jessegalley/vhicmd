package responseparser

func getPowerStateString(state int) string {
	switch state {
	case 0:
		return "NOSTATE"
	case 1:
		return "RUNNING"
	case 3:
		return "PAUSED"
	case 4:
		return "SHUTDOWN"
	case 6:
		return "CRASHED"
	case 7:
		return "SUSPENDED"
	default:
		return "UNKNOWN"
	}
}

// Helper function
func stringOrNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
