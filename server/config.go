package server

type ConfigFormatInterface interface {
	TransferToRegistry() *Registry
	TransferToFormat(*Registry)
	Write(string) error
	Parse(string) error
	ParseFromText(string) error
}

type Config struct {
	Registry *Registry
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
	c.Registry = c.Format.TransferToRegistry()
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
	c.Registry = c.Format.TransferToRegistry()
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