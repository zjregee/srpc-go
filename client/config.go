package client

type ConfigFormatInterface interface {
	TransferToNetWork() *Network
	TransferToFormat(*Network)
	Write(string) error
	Parse(string) error
	ParseFromText(string) error
}

type Config struct {
	Network *Network
	Format  ConfigFormatInterface
}

func NewConfig(confName string, format ConfigFormatInterface) (*Config, error) {
	c := &Config{
		Format: format,
	}
	err := c.Parse(confName)
	if err != nil {
		return nil, err
	}
	c.Network = c.Format.TransferToNetWork()
	return c, nil
}

func NewConfigFromText(text string, format ConfigFormatInterface) (*Config, error) {
	c := &Config{
		Format: format,
	}
	err := c.ParseFromText(text)
	if err != nil {
		return nil, err
	}
	c.Network = c.Format.TransferToNetWork()
	return c, nil
}

func (c *Config) Parse(fname string) (err error) {
	return c.Format.Parse(fname)
}

func (c *Config) ParseFromText(text string) error {
	return c.Format.ParseFromText(text)
}

func (c *Config) Write(fname string) (err error) {
	return c.Format.Write(fname)
}