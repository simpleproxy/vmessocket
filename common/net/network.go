package net

func HasNetwork(list []Network, network Network) bool {
	for _, value := range list {
		if value == network {
			return true
		}
	}
	return false
}

func (n Network) SystemString() string {
	switch n {
	case Network_TCP:
		return "tcp"
	case Network_UDP:
		return "udp"
	case Network_UNIX:
		return "unix"
	default:
		return "unknown"
	}
}
