package ref_test

import (
	"fmt"
	"time"

	"github.com/satori/go.uuid"
	"github.com/shopspring/decimal"

	"github.com/jiarung/mochi/common/ref"
)

func ExamplePlainText() {
	type myType struct {
		ID   uuid.UUID ``
		Name string    ``
		cc   string    ``
		Date time.Time `sign:"-"`
		Note string    `sign:"-"`
	}

	m := myType{
		ID:   uuid.NewV4(),
		Name: "Foo",
		cc:   "Bar",
		Date: time.Now(),
		Note: "go rough",
	}

	p, err := ref.PlainText(&m)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(p))

	// It would generate a plain text such as:
	//
	//	model-signature ID:"5a147d3b-389d-4eba-a296-1d4f147e7205" Name:"Foo"
	//
	// The unexposed field cc and tagged sign:"-" fields Date, Note would not
	// be writed into the plain text.
}

func Example_rawMarshalAndUnMarshal() {
	type Base struct {
		UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp with time zone" json:"-" sign:"-"`
		CreatedAt time.Time `gorm:"column:created_at;type:timestamp with time zone" json:"-" sign:"-"`
	}
	type Ledger struct {
		Base

		ID          uuid.UUID       `gorm:"column:id;primary_key;type:uuid;not null;default:uuid_generate_v4()" json:"-"`
		Timestamp   time.Time       `gorm:"column:timestamp;type:timestamp with time zone;not null" json:"timestamp"`
		Amount      decimal.Decimal `gorm:"column:amount;type:decimal(38,18);not null" json:"amount"`
		Balance     decimal.Decimal `gorm:"column:balance;type:decimal(38,18);not null" json:"balance"`
		Description string          `gorm:"column:description;type:varchar(256)" json:"-"`
		TradeID     *uuid.UUID      `gorm:"column:trade_id;type:uuid" json:"trade_id"`
		DepositID   *uuid.UUID      `gorm:"column:deposit_id;type:uuid" json:"deposit_id"`
	}

	tradeID := uuid.NewV4()
	l := Ledger{
		Base: Base{
			UpdatedAt: time.Now().UTC(),
			CreatedAt: time.Now().UTC(),
		},
		ID:          uuid.NewV4(),
		Timestamp:   time.Now().UTC(),
		Amount:      decimal.NewFromFloat(1011.13),
		Balance:     decimal.Zero,
		Description: "We need go rough",
		TradeID:     &tradeID,
	}

	encoded, err := ref.RawMarshal(l)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(encoded))

	var decoded Ledger
	err = ref.RawUnmarshal(encoded, &decoded)
	if err != nil {
		fmt.Println(err)
	}
}
