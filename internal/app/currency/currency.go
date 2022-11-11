package currency

type Currency struct {
	v string
}

var (
	UNKNOWN = Currency{"UNKNOWN"}
	RUR     = Currency{"RUR"}
	GEL     = Currency{"GEL"}
	AMD     = Currency{"AMD"}
	USD     = Currency{"USD"}
	EUR     = Currency{"EUR"}
)

func (c Currency) String() string {
	return c.v
}

func FromString(value string) (Currency, error) {
	switch value {
	case RUR.String():
		return RUR, nil
	case GEL.String():
		return GEL, nil
	case AMD.String():
		return AMD, nil
	case USD.String():
		return USD, nil
	case EUR.String():
		return EUR, nil
	}

	return UNKNOWN, nil
}
