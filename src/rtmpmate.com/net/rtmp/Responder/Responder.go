package Responder

import ()

type Responder struct {
	ID     int
	Result func()
	Status func()
}

func New(id int, result func(), status func()) (*Responder, error) {
	var res Responder
	res.ID = id
	res.Result = result
	res.Status = status

	return &res, nil
}
