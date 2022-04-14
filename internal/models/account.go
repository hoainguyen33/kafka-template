package models

import (
	accountsService "kafka-test/proto/account"
)

type Account struct {
	Username string `json:"username" bson:"username,omitempty"`
	Password string `json:"password,omitempty" bson:"password,omitempty"`
}

func (p *Account) ToProto() *accountsService.Account {
	return &accountsService.Account{
		Username: p.Username,
		Password: p.Password,
	}
}

func (p *Account) ToTokenProto() string {
	// token
	return "e7c59c2ac1930cde78b4720919233b8d42d96537"
}
