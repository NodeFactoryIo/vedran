package models

type Node struct {
	ID            string `storm:"id"`
	ConfigHash    string
	NodeUrl       string
	PayoutAddress string
	Token         string
	Cooldown      int64
	LastUsed      int64
}

