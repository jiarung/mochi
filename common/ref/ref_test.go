package ref

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cobinhood/mochi/types"
)

// Base makes a local definition instead of import
// github.com/cobinhood/mochi/models to avoid from circle import.
type Base struct {
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp with time zone" json:"-" sig:"-"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp with time zone" json:"-" sig:"-"`
}

// SignedModelBase makes a local definition instead of import
// github.com/cobinhood/mochi/models to avoid from circle import.
type SignedModelBase struct {
	Signature string `gorm:"column:sig;type:varchar(64)" json:"-" sig:"-"`
}

// Ledger makes a local definition instead of import
// github.com/cobinhood/mochi/models to avoid from circle import.
type Ledger struct {
	Base
	SignedModelBase

	ID               uuid.UUID          `gorm:"column:id;primary_key;type:uuid;not null;default:uuid_generate_v4()" json:"-"`
	Count            int                `gorm:"column:count;type:int" json:"count"`
	Timestamp        time.Time          `gorm:"column:timestamp;type:timestamp with time zone;not null" json:"timestamp"`
	UserID           uuid.UUID          `gorm:"column:user_id;type:uuid;not null" json:"-"`
	CurrencyID       string             `gorm:"column:currency_id;type:varchar(16);not null" json:"currency"`
	Type             types.LedgerType   `gorm:"column:type;type:varchar(64);not null" json:"type"`
	Action           types.LedgerAction `gorm:"column:action;type:varchar(64);not null" json:"action"`
	Amount           decimal.Decimal    `gorm:"column:amount;type:decimal(38,18);not null" json:"amount"`
	Balance          decimal.Decimal    `gorm:"column:balance;type:decimal(38,18);not null" json:"balance"`
	Description      string             `gorm:"column:description;type:varchar(256)" json:"-"`
	TradeID          *uuid.UUID         `gorm:"column:trade_id;type:uuid" json:"trade_id"`
	DepositID        *uuid.UUID         `gorm:"column:deposit_id;type:uuid" json:"deposit_id"`
	WithdrawalID     *uuid.UUID         `gorm:"column:withdrawal_id;type:uuid" json:"withdrawal_id"`
	FiatDepositID    *uuid.UUID         `gorm:"column:fiat_deposit_id;type:uuid" json:"fiat_deposit_id"`
	FiatWithdrawalID *uuid.UUID         `gorm:"column:fiat_withdrawal_id;type:uuid" json:"fiat_withdrawal_id"`
}

type refTestSuit struct {
	suite.Suite
}

func (s *refTestSuit) TestRawMarshalUnmarshal() {
	tradeID := uuid.NewV4()
	l := Ledger{
		Base: Base{
			UpdatedAt: time.Now().UTC(),
			CreatedAt: time.Now().UTC(),
		},
		SignedModelBase: SignedModelBase{
			Signature: "hello wo\"rld",
		},
		ID:          uuid.NewV4(),
		Count:       5566,
		Timestamp:   time.Now().UTC(),
		UserID:      uuid.NewV4(),
		CurrencyID:  "USDT",
		Type:        types.LedgerTypeFundingLedger,
		Action:      types.ActionDeposit,
		Amount:      decimal.NewFromFloat(1011.13),
		Balance:     decimal.Zero,
		Description: "We need go rough",
		TradeID:     &tradeID,
	}

	// Marshal struct to string.
	encoded, err := RawMarshal(l)
	require.Nil(s.T(), err)
	require.True(s.T(), strings.Contains(string(encoded), "UserID"))

	fmt.Println("Marshal#1:", string(encoded))

	// Unmahrshal string to struct.
	var decoded Ledger
	err = RawUnmarshal(encoded, &decoded)
	require.Nil(s.T(), err)
	require.Equal(s.T(), l.UpdatedAt, decoded.UpdatedAt)
	require.Equal(s.T(), l.Timestamp, decoded.Timestamp)
	require.Equal(s.T(), l.Signature, decoded.Signature)
	require.Equal(s.T(), l.Type, decoded.Type)
	require.Equal(s.T(), l.WithdrawalID, decoded.WithdrawalID)
	require.Equal(s.T(), l.Count, decoded.Count)
	require.True(s.T(), uuid.Equal(*l.TradeID, *decoded.TradeID))
	require.True(s.T(), uuid.Equal(l.ID, decoded.ID))
	require.True(s.T(), uuid.Equal(l.UserID, decoded.UserID))
	require.True(s.T(), l.Amount.Equal(decoded.Amount))

	// Marshal the decoded one to string and compare with the 1st JSON string.
	encoded2, err := RawMarshal(decoded)
	require.Nil(s.T(), err)
	// Equal means no data lost during marshal-unmarshal.
	require.Equal(s.T(), encoded, encoded2)

	fmt.Println("Marshal#2:", string(encoded2))
}

func (s *refTestSuit) TestHelpFunction() {
	tradeID := uuid.NewV4()
	l := Ledger{
		Base: Base{
			UpdatedAt: time.Now().UTC(),
			CreatedAt: time.Now().UTC(),
		},
		SignedModelBase: SignedModelBase{
			Signature: "hello world",
		},
		ID:          uuid.NewV4(),
		Timestamp:   time.Now().UTC(),
		UserID:      uuid.NewV4(),
		CurrencyID:  "USDT",
		Type:        types.LedgerTypeFundingLedger,
		Action:      types.ActionDeposit,
		Amount:      decimal.NewFromFloat(1011.13),
		Balance:     decimal.Zero,
		Description: "We need go rough",
		TradeID:     &tradeID,
	}

	fmt.Println(BeautifyStruct(l, true))
	fmt.Println("Fields:", ListField(l))
	fmt.Println("Methods:", ListMethod(uuid.NewV4()))
}

func (s *refTestSuit) TestGenPlainText() {
	type myType struct {
		ID   uuid.UUID ``
		Name string    ``
		cc   string    ``
		Date time.Time `sig:"-"`
		Note string    `sig:"-"`
	}

	m := myType{
		ID:   uuid.NewV4(),
		Name: "POPO",
		cc:   "dd",
		Date: time.Now(),
		Note: "go rough",
	}

	p, err := PlainText(&m)
	require.Nil(s.T(), err)

	expected := fmt.Sprintf(`model-signature ID:"%v" Name:"%v"`, m.ID, m.Name)
	require.Equal(s.T(), string(p), expected)

	fmt.Println("PlaintTex:", string(p))
}

func TestRef(test *testing.T) {
	suite.Run(test, &refTestSuit{})
}
