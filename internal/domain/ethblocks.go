package domain

type EthBlocks struct {
	Data struct {
		Blocks []Block `json:"blocks"`
	} `json:"data"`
}

type Block struct {
	ID string `json:"id"`
}
