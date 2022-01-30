//go:build !confonly
// +build !confonly

package inbound

func (c *Config) GetDefaultValue() *DefaultConfig {
	if c.GetDefault() == nil {
		return &DefaultConfig{}
	}
	return c.Default
}
