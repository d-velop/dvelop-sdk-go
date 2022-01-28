package jsonlog

type Logdata struct {
	Name       string
	Visibility int
	Attributes *Attributes
}

func NewLogdata() Logdata {
	logdata := Logdata{}
	logdata.Visibility = 1
	return logdata
}

func (l *Logdata) AddToEvent(e *Event) {
	e.Visibility = &l.Visibility
	if l.Name != "" {
		e.Name = l.Name
	}
	if l.Attributes != nil {
		e.Attributes = l.Attributes
	}
}
