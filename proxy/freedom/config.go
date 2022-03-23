package freedom

func (c *Config) useIP() bool {
	return c.DomainStrategy == Config_USE_IP
}
