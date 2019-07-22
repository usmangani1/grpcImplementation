package types

type Order struct {
	Name string
	Asc  bool
}

type Page struct {
	Page   int64
	Limit  int64
	Orders *[]Order
}

type Pageable struct {
	Data     interface{} `json:"data"`
	Count    int64       `json:"count"`
	PageSize int64       `json:"pageSize"`
}
