/*

Package customquery allows user to create custom SQL with JSON.

The main interface of this package is the `SQLOptions` struct that could apply
custom options to a gorm db query. SQLOptions has "filter", "order", "limit"
and "page" fields, which will be explained below.

Usage

Take "get_kycs" endpoints for example, which allows user to query kyc table with
filters and orders.
- JSON Request body
    {
        "filter":{...},
        "order":[{...}],
        "limit": 20,
        "page": 3
    }

Create "SQLOptions" instance from post body of request
    opt, err := customquery.NewSQLOptions(ginCtx, nil, nil)

Apply "SQLOptions" to a DB query, in this case
	db.Model(&models.KYC{})
, with
	defaultLimit = 50
	defaultPage = 1
. The "defaultLimit" and "defaultPage" parameters will be override if "Limit"
and "Page" is specified in "SQLOptions".
Since the functions call `Count` query first to find out `totalPage`, there
might be a error raised in the query.
    db, limit, page, totalPage, err := opt.Apply(db.Model(&models.KYC{}), 50, 1, 0)

Execute the query
    var kycs []models.KYC
    db.Find(&kycs)

Filter

Filter is a nested map that will be evaluated to `Where` arguments in a SQL query.

Each layer is a single-key map, with operator as the key and parameters as
the value. There are two kinds of filters (operators), comparison filters and
logic filters. Comparison filter is the atomic element of a filter and logic
filter is to combine multiple filters into one single filter

Comparison Operators:
	- equal
	- not_equal
	- greater_than
	- greater_than_or_equal
	- smaller_than
	- smaller_than_or_equal
	- in
	- like
	- between

Logic Operators:
	- and
	- or

Examples

ex. 1: "Where column_a = 'A'"
	{
		"equal": {
			"column": "column_a",
			"value": "A"
		}
	}

ex. 2: "Where ((column_a = 'A') AND (column_b != 'B'))"
	{
		"and": [
			{
				"equal": {
					"column": "column_a",
					"value": "A"
				}
			},{
				"not_equal": {
					"column": "column_b",
					"value": "A"
				}
			}
		]
	}

Filter Validator

FilterComparisonValidator is called to validate comparison operation, which
could prevent unwanted queries.

	var v = func(column string, operator string, value interfaces{}) error {
		if column == "secret_column" {
			return errors.New("attempt to filter on secret column")
		}
		return nil
	}

Order

Order is a struct that will be evaluated to `ORDER BY` arguments in a SQL query.
	- Column:  is the column to order
	- Keyword:  is the way rows will be ordered. (`asc`|`desc`)

Order Validator

OrderValidator is called to validate order operation, which could prevent
unwanted queries.
	var v = func(column, keyword string) error {
		if column == "secret_column" {
			return errors.New("attempt to order by secret column")
		}
		return nil
	}
*/
package customquery
