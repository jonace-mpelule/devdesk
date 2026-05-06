package devdesk

type stringList []string

func (s *stringList) String() string {
	return ""
}

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}
