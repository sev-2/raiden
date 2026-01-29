package postgres

const (
	Select    = "SELECT"
	From      = "FROM"
	Where     = "WHERE"
	Order     = "ORDER"
	OrderBy   = "ORDER BY"
	Limit     = "LIMIT"
	Offset    = "OFFSET"
	Group     = "GROUP"
	GroupBy   = "GROUP BY"
	Having    = "HAVING"
	Insert    = "INSERT"
	Update    = "UPDATE"
	Delete    = "DELETE"
	Create    = "CREATE"
	Alter     = "ALTER"
	Drop      = "DROP"
	Truncate  = "TRUNCATE"
	Join      = "JOIN"
	Inner     = "INNER"
	Outer     = "OUTER"
	Left      = "LEFT"
	Right     = "RIGHT"
	LeftJoin  = "LEFT JOIN"
	RightJoin = "RIGHT JOIN"
	InnerJoin = "INNER JOIN"
	OuterJoin = "OUTER JOIN"
	On        = "ON"
	As        = "AS"
	And       = "AND"
	Or        = "OR"
	Not       = "NOT"
	Between   = "BETWEEN"
	In        = "IN"
	Like      = "LIKE"
	Exists    = "EXISTS"
	All       = "ALL"
	Any       = "ANY"
	Union     = "UNION"
	Intersect = "INTERSECT"
	Except    = "EXCEPT"
	Asc       = "ASC"
	Desc      = "DESC"
	Is        = "IS"
	IsNull    = "IS NULL"
	IsNotNull = "IS NOT NULL"
	Case      = "CASE"
	When      = "WHEN"
	Then      = "THEN"
	Else      = "ELSE"
	End       = "END"
	With      = "WITH"
)

// ReservedKeywords contains all reserved keywords in PostgreSQL
var ReservedKeywords = map[string]struct{}{
	Select:    {},
	From:      {},
	Where:     {},
	Order:     {},
	OrderBy:   {},
	Limit:     {},
	Offset:    {},
	Group:     {},
	GroupBy:   {},
	Having:    {},
	Insert:    {},
	Update:    {},
	Delete:    {},
	Create:    {},
	Alter:     {},
	Drop:      {},
	Truncate:  {},
	Join:      {},
	Inner:     {},
	Outer:     {},
	Left:      {},
	Right:     {},
	LeftJoin:  {},
	RightJoin: {},
	InnerJoin: {},
	OuterJoin: {},
	On:        {},
	As:        {},
	And:       {},
	Or:        {},
	Not:       {},
	Between:   {},
	In:        {},
	Like:      {},
	Exists:    {},
	All:       {},
	Any:       {},
	Union:     {},
	Intersect: {},
	Except:    {},
	Asc:       {},
	Desc:      {},
	Is:        {},
	IsNull:    {},
	IsNotNull: {},
	Case:      {},
	When:      {},
	Then:      {},
	Else:      {},
	End:       {},
	With:      {},
}

func IsReservedKeyword(str string) bool {
	_, ok := ReservedKeywords[str]
	return ok
}

var symbols = map[string]struct{}{
	"=":   {},
	"<>":  {},
	"!=":  {},
	">":   {},
	"<":   {},
	">=":  {},
	"<=":  {},
	"+":   {},
	"-":   {},
	"*":   {},
	"/":   {},
	"(":   {},
	")":   {},
	",":   {},
	";":   {},
	".":   {},
	":":   {},
	"::":  {},
	"::=": {},
	"||":  {},
	"<=>": {},
	"&":   {},
	"|":   {},
	"^":   {},
	"<<":  {},
	">>":  {},
	"~":   {},
	"<<=": {},
	">>=": {},
	"&=":  {},
	"|=":  {},
	"^=":  {},
	"~=":  {},
	"%":   {},
	"@":   {},
	"#":   {},
	"$":   {},
	"`":   {},
	"[":   {},
	"]":   {},
	"{":   {},
	"}":   {},
	"!":   {},
	"?":   {},
	":=":  {},
	"=>":  {},
}

func IsReservedSymbol(str string) bool {
	_, ok := symbols[str]
	return ok
}
