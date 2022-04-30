package config

type ServerSettingS struct {
	Port string
}

type EmailSettingS struct {
	ServerHost   string
	ServerPort   int
	FromEmail    string
	FromPassword string
}

func (s *Setting) ReadSection(k string, v interface{}) error {
	err := s.vp.UnmarshalKey(k, v)
	if err != nil {
		return err
	}

	return nil
}
