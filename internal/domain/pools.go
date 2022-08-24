package domain

type Pools struct {
	Data struct {
		Pools []Pool `json:"pools"`
	} `json:"data"`
}

type Pool struct {
	ID                  string `json:"id"`
	FeeTier             string `json:"feeTier"`
	VolumeUSD           string `json:"volumeUSD"`
	TotalValueLockedUSD string `json:"totalValueLockedUSD"`
}
