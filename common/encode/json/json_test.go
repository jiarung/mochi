package json

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

type testJSONSuite struct {
	suite.Suite
}

func (s *testJSONSuite) TestObj() {
	{
		obj := Object()
		res, err := obj.Marshal()
		s.Require().NoError(err)
		s.Require().Equal("{}", string(res))
	}

	{
		obj := Object(Attr("a", "1"))
		res, err := obj.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(`{"a":"1"}`, string(res))
	}

	{
		subObj := Object()
		obj := Object(Attr("sub", subObj))
		res, err := obj.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(`{"sub":{}}`, string(res))
	}

	{
		obj := Object(Attr("a", 1), Attr("b", 2), Attr("time", nil))
		res, err := obj.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(`{"a":1,"b":2,"time":null}`, string(res))
	}

	{
		obj := Object(Attr("a", 1), Attr("b", 0.4), Attr("enable", false))
		res, err := obj.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(`{"a":1,"b":0.4,"enable":false}`, string(res))
	}
}

func (s *testJSONSuite) TestObjAttr() {
	{
		obj := Object(Attr("a", Object()))
		res, err := obj.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(`{"a":{}}`, string(res))
	}

	{
		obj := Object(Attr("Q",
			Object(
				Attr("a", 1),
				Attr("b", 2),
			),
		))
		res, err := obj.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(`{"Q":{"a":1,"b":2}}`, string(res))
	}

	{
		obj := Object(Attr("a", Attr("b", Attr("c", Attr("d", Object())))))
		res, err := obj.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(`{"a":{"b":{"c":{"d":{}}}}}`, string(res))
	}

	{
		obj := Object(Attr("a", Array(Object(Attr("b", Array(1, Object(), "c"))))))
		res, err := obj.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(`{"a":[{"b":[1,{},"c"]}]}`, string(res))
	}
}

func (s *testJSONSuite) TestObjAppend() {
	{
		obj := Object(Attr("a", "A"), Attr("b", 99), Attr("c", "C"))
		obj2 := Object()
		obj2.Append(Attr("a", "A"))
		obj2.Append(Attr("b", 99), Attr("c", "C"))
		res, err := obj.Marshal()
		res2, err2 := obj2.Marshal()
		s.Require().NoError(err)
		s.Require().NoError(err2)

		s.Require().Equal(res, res2)
	}
}

func (s *testJSONSuite) TestArray() {
	{
		arr := Array()
		res, err := arr.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(`[]`, string(res))
	}

	{
		arr := Array("a", "b", "c")
		res, err := arr.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(`["a","b","c"]`, string(res))
	}

	{
		arr := Array()
		arr.Append("a", "b")
		arr.Append(nil)
		arr.Append(false)
		res, err := arr.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(`["a","b",null,false]`, string(res))
	}

	{
		arr := Array()
		arr.Append("a", "b")
		arr.Append(3)
		res, err := arr.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(`["a","b",3]`, string(res))
	}
}

func (s *testJSONSuite) TestArrayAppendAny() {
	{
		arr := Array()
		arr.Append(Object(), Object(Attr("a", "b")))
		res, err := arr.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(`[{},{"a":"b"}]`, string(res))
	}

	{
		arr := Array()
		arr.Append(Object(), Array())
		res, err := arr.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(`[{},[]]`, string(res))
	}
}

func (s *testJSONSuite) TestComplexObject() {
	{
		user := Object()
		user.Append(Attr("user_id", "3345678"))
		user.Append(Attr("name", "xnum"))
		user.Append(Attr("country", "tw"))

		location := Array()
		location.Append(124.012341, 23.998745)

		event := Object(
			Attr("action", "login"),
			Attr("service", "auth"),
		)

		payload := Object(
			Attr("user", user),
			Attr("loc", location),
			Attr("event", event),
		)
		data, err := payload.Marshal()
		s.Require().NoError(err)

		event.Append(Attr("source", "gmail"))
		data2, err := payload.Marshal()
		s.Require().NoError(err)

		s.Require().NotEqual(string(data), string(data2))
	}

	{
		alice := Object(
			Attr("name", "alice"),
			Attr("class", "1-A"),
			Attr("recent_score", []float64{98.5, 99.7, 100.0}),
		)

		bob := Object(
			Attr("name", "bob"),
			Attr("class", "1-B"),
			Attr("recent_score", []float64{18.5, 99.3, 12.3}),
		)

		stus := Array(alice, bob)
		res, err := stus.Marshal()
		s.Require().NoError(err)
		s.Require().Equal(`[{"class":"1-A","name":"alice","recent_score":[98.5,99.7,100]},{"class":"1-B","name":"bob","recent_score":[18.5,99.3,12.3]}]`, string(res))
	}

}

func (s *testJSONSuite) TestIllegalCase() {
	attr := Attr("name", "xnum")
	_, err := json.Marshal(attr)
	s.Require().Error(err)

	// test passing address
	_, err = json.Marshal(&attr)
	s.Require().Error(err)

	arr := Array(Attr("name", "xnum"))
	_, err = arr.Marshal()
	s.Require().Error(err)

	obj := Object(Attr("arr", arr))
	_, err = obj.Marshal()
	s.Require().Error(err)
}

func TestJson(t *testing.T) {
	suite.Run(t, &testJSONSuite{})
}
