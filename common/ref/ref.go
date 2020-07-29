package ref

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"
)

// RawMarshal marshalls a struct to JSON string without refer to json tag.
// That is, you can transfer the whole struct to string evan if you assign
// json:"-" tag to some field for some purpose.
//
// Because it doesn't refer to json tag, it uses the original field name as the
// parameter name in JSON string, not the familiar underscore_case of JSON.
// Such as:
//
//	{"UpdatedAt":"2017-12-15T09:29:10.132082962Z","UserID":"eb2e2ac5-6b09-42f3-9929-ef9cc1bfab3a"}
//
// This function would skip marshalling the field which has the gorm ForeignKey
// tag because the tag means it is a reference to other struct, not a native
// field of this strcut.
//
// 	gorm:"ForeignKey:"
//
// Besides, this function would skip all the unexposed field.
func RawMarshal(d interface{}) ([]byte, error) {
	var s []string
	if err := recursiveEncode(d, &s); err != nil {
		return []byte(""), err
	}
	output := strings.Join(s, ",")

	return []byte("{" + output + "}"), nil
}

func recursiveEncode(d interface{}, s *[]string) error {
	v := reflect.ValueOf(d)
	t := reflect.TypeOf(d)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		return errors.New("not a struct")
	}

	for i := 0; i < t.NumField(); i++ {
		if !v.Field(i).CanInterface() {
			continue
		}
		if t.Field(i).Anonymous {
			// It is an embeded structure, recursive passing it.
			if err := recursiveEncode(v.Field(i).Interface(), s); err != nil {
				return err
			}
		} else if !strings.Contains(t.Field(i).Tag.Get("gorm"), "ForeignKey:") {
			v, _ := json.Marshal(v.Field(i).Interface())
			value, _ := json.Marshal(string(v))
			*s = append(*s, fmt.Sprintf(`"%v":%v`, t.Field(i).Name, string(value)))
		}
	}

	return nil
}

// RawUnmarshal unmarshalls JSON string to structure without refer to json tag.
// Same as RawMarshal, this function would skip all the unexposed fields & gorm
// ForeignKey.
func RawUnmarshal(s []byte, d interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(d))

	var strMap map[string]string
	if err := json.Unmarshal(s, &strMap); err != nil {
		return err
	}

	return recursiveDecode(strMap, v)
}

func recursiveDecode(strMap map[string]string, v reflect.Value) error {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanInterface() {
			continue
		}
		if t.Field(i).Anonymous {
			// It is an embeded structure, recursive passing it.
			if err := recursiveDecode(strMap, v.Field(i)); err != nil {
				return err
			}
		} else if !strings.Contains(t.Field(i).Tag.Get("gorm"), "ForeignKey:") {
			value, ok := strMap[t.Field(i).Name]
			if !ok {
				return fmt.Errorf("key %v is not found", t.Field(i).Name)
			}
			// reflect.New returns a pointer type so it's no need to fetch the
			// address by prependinf '&' to assign to json.Unmarshal().
			newValue := reflect.New(t.Field(i).Type)
			if t.Field(i).Type.Kind() != reflect.Ptr || value != "" {
				// In JSON, some types needs quote sign but some doesn't need,
				// so here attempts to unmarshal both cases: with & without quote.
				// Eventually one of two cases would be successed.
				instance := newValue.Interface()
				err := json.Unmarshal([]byte(value), instance)
				if err != nil {
					if err = json.Unmarshal(
						[]byte(`"`+value+`"`), instance); err != nil {
						return err
					}
				}
			}
			// newValue is a pointer so using Indirect() to get its value.
			v.Field(i).Set(reflect.Indirect(newValue))
		}
	}

	return nil
}

// PlainText generates the plain text from the given struct for signing model
// signature. It would skip the field with tag
//
//	`sig="-"`
//
// The output string would be in the form:
//
//	model-signature ID:"5a147d3b-389d-4eba-a296-1d4f147e7205" Name:"Foo"
//
// model-signature is a constant leading key word, followed by each signable
// field:"value" pairs.
//
// Same as RawMarshal, this function would skip all the unexposed fields & gorm
// ForeignKey.
func PlainText(d interface{}) ([]byte, error) {
	// !!!!!! MOST IMPORTANT !!!!!!
	// DO NOT change the following line otherwise it would cause
	// validate model-signature failed.
	s := "model-signature"
	// !!!!!! MOST IMPORTANT !!!!!!
	if err := recursiveGenPlainText(d, &s); err != nil {
		return []byte(""), err
	}

	return []byte(s), nil
}

func recursiveGenPlainText(d interface{}, s *string) error {
	v := reflect.ValueOf(d)
	t := reflect.TypeOf(d)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		return errors.New("not a struct")
	}

	for i := 0; i < t.NumField(); i++ {
		if !v.Field(i).CanInterface() {
			continue
		}
		if t.Field(i).Anonymous {
			// It is an embeded structure, recursive passing it.
			err := recursiveGenPlainText(v.Field(i).Interface(), s)
			if err != nil {
				return err
			}
		} else if t.Field(i).Tag.Get("sig") != "-" &&
			!strings.Contains(t.Field(i).Tag.Get("gorm"), "ForeignKey:") {
			var value []byte
			var err error
			switch t := v.Field(i).Interface(); t.(type) {
			case *time.Time:
				if v.Field(i).IsNil() {
					value = []byte("nil")
				} else {
					value = []byte(fmt.Sprintf("%d", t.(*time.Time).Unix()))
				}
			case time.Time:
				value = []byte(fmt.Sprintf("%d", t.(time.Time).Unix()))
			default:
				value, err = json.Marshal(v.Field(i).Interface())
				if err != nil {
					return err
				}
			}

			// !!!!!! MOST IMPORTANT !!!!!!
			// DO NOT change the following line otherwise it would cause
			// validate model-signature failed.
			*s = *s + fmt.Sprintf(` %v:%v`, t.Field(i).Name, string(value))
			// !!!!!! MOST IMPORTANT !!!!!!
		}
	}

	return nil
}

// ListMethod list all the methods of the given variable.
func ListMethod(v interface{}) []string {
	var s []string

	t := reflect.TypeOf(v)
	for i := 0; i < t.NumMethod(); i++ {
		s = append(s, fmt.Sprintf("%v", t.Method(i).Name))
	}

	return s
}

// ListField lists all the fields of the given structure.
func ListField(v interface{}) []string {
	var s []string

	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			s = append(s, fmt.Sprintf("%v", t.Field(i).Name))
		}
	}

	return s
}

// BeautifyStruct shows the given structure in the beautiful form:
//
//	{
//		UpdatedAt	Time	2017-12-15 09:29:10.131804303 +0000 UTC
//		CreatedAt	Time	2017-12-15 09:29:10.131804342 +0000 UTC
//		Signature	string	hello world
//		ID	UUID	f0776b47-bd19-4fa7-aa5c-e3f629ee4846
//		Timestamp	Time	2017-12-15 09:29:10.131804507 +0000 UTC
//		UserID	UUID	9c1a4521-4add-4da7-b579-e5922f5a8031
//		CurrencyID	string	USD
//		Type	LedgerType	funding
//		Action	LedgerAction	deposit
//		Amount	Decimal	1011.13
//		Balance	Decimal	0
//		Description	string	We need go rough
//		TradeID	*UUID	2c5b861c-8438-469a-a962-66d499de951b
//		DepositID	*UUID	<nil>
//		WithdrawalID	*UUID	<nil>
//		FiatDepositID	*UUID	<nil>
//		FiatWithdrawalID	*UUID	<nil>
//	}
//
// Assisgh attribute color=true would make the output colorful.
//
// You can rewrite the String() method of your struct to print it easily.
func BeautifyStruct(d interface{}, color bool) string {
	var s []string
	if err := recursiveBeautify(d, &s, color); err != nil {
		fmt.Println(err)
		return ""
	}
	output := strings.Join(s, "\n")

	return "{\n" + output + "\n}"
}

func recursiveBeautify(d interface{}, s *[]string, color bool) error {
	v := reflect.ValueOf(d)
	t := reflect.TypeOf(d)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	if t.Kind() != reflect.Struct {
		return errors.New("not a struct")
	}

	maxField := 0
	maxType := 0
	for i := 0; i < t.NumField(); i++ {
		maxField = int(math.Max(
			float64(maxField),
			float64(len(t.Field(i).Name))))
		maxType = int(math.Max(
			float64(maxType),
			float64(len(t.Field(i).Type.String()))))
	}

	maxFieldStr := fmt.Sprintf("%d", maxField+2)
	maxTypeStr := fmt.Sprintf("%d", maxType+2)

	for i := 0; i < t.NumField(); i++ {
		if !v.Field(i).CanInterface() {
			continue
		}
		printFmt := "  %" + maxFieldStr + "s  %-" + maxTypeStr + "s  %v"
		if color {
			printFmt = "  \x1b[36m%" + maxFieldStr + "s\x1b[m  \x1b[32m%-" + maxTypeStr + "s\x1b[m  \x1b[33m%v\x1b[m"
		}

		fieldStr := t.Field(i).Name
		typeStr := t.Field(i).Type.String()
		if t.Field(i).Type.Kind() == reflect.Ptr {
			typeStr = "*" + t.Field(i).Type.Elem().String()
		} else if t.Field(i).Type.Kind() == reflect.Slice {
			typeStr = "[]" + t.Field(i).Type.Elem().String()
		}
		typeStr = strings.Replace(typeStr, "uint8", "byte", 1)

		if t.Field(i).Anonymous {
			// It is an embeded structure, recursive passing it.
			err := recursiveBeautify(v.Field(i).Interface(), s, color)
			if err != nil {
				return err
			}
		} else if !strings.Contains(t.Field(i).Tag.Get("gorm"), "ForeignKey:") {
			*s = append(*s, fmt.Sprintf(
				printFmt, fieldStr, typeStr, v.Field(i)))
		}
	}

	return nil
}
