package coa

import (
	"time"

	"github.com/google/uuid"
)

type AccountType string

const (
	Asset     AccountType = "ASSET"
	Liability AccountType = "LIABILITY"
	Equity    AccountType = "EQUITY"
	Revenue   AccountType = "REVENUE"
	COGS      AccountType = "COGS"
	Expense   AccountType = "EXPENSE"
)

type NormalBalance string

const (
	Debit  NormalBalance = "DEBIT"
	Credit NormalBalance = "CREDIT"
)

type Account struct {
	ID            uuid.UUID     `json:"id"`
	Code          string        `json:"code"`
	Name          string        `json:"name"`
	AccountType   AccountType   `json:"account_type"`
	NormalBalance NormalBalance `json:"normal_balance"`
	ParentID      *uuid.UUID    `json:"parent_id,omitempty"`
	IsPostable    bool          `json:"is_postable"`
	IsActive      bool          `json:"is_active"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type CreateAccountInput struct {
	Code          string      `json:"code"`
	Name          string      `json:"name"`
	AccountType   AccountType `json:"account_type"`
	NormalBalance NormalBalance `json:"normal_balance"`
	ParentID      *uuid.UUID  `json:"parent_id,omitempty"`
	IsPostable    *bool       `json:"is_postable,omitempty"`
}

type UpdateAccountInput struct {
	Name       string `json:"name"`
	IsPostable bool   `json:"is_postable"`
	IsActive   bool   `json:"is_active"`
}
