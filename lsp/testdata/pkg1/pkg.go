package pkg1

type Language struct {
	Name string
}

func (l *Language) String() string {
	return l.Name
}
