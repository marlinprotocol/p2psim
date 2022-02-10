package episub

// Source: https://asc.di.fct.unl.pt/~jleitao/pdf/dsn07-leitao.pdf

var (
	Crand = 3
	Cnear = 4
	K     = 6
	ARWL  = 5
	PRWL  = 3
)

type MembershipConfig struct {
	Crand *int `toml:"c_rand,omitempty"`
	Cnear *int `toml:"c_near,omitempty"`
	K     *int `toml:"k,omitempty"`
	ARWL  *int `toml:"ARWL,omitempty"`
	PRWL  *int `toml:"PRWL,omitempty"`
}

func GetDefaultConfig() *MembershipConfig {
	return &MembershipConfig{
		Crand: &Crand,
		Cnear: &Cnear,
		K:     &K,
		ARWL:  &ARWL,
		PRWL:  &PRWL,
	}
}
