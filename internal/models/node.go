package models

type Node struct {
	ID            string `storm:"id"`
	ConfigHash    string
	PayoutAddress string
	Token         string
	Cooldown      int
	LastUsed      int64
}

