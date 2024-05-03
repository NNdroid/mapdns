package http_srv

type Resp struct {
	Status bool        `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

func NewRespOK(data interface{}) *Resp {
	return &Resp{
		Status: true,
		Msg:    "ok",
		Data:   data,
	}
}

func NewRespFail(data interface{}) *Resp {
	return &Resp{
		Status: false,
		Msg:    "fail",
		Data:   data,
	}
}
