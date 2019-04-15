package respond

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
)

type Base struct {
	r        *http.Request
	err      error
	reqBody  []byte
	respBody []byte
	step     int
}

func (b *Base) Error() error {
	return b.err
}

func (b *Base) CustomFn(f func() error) *Base {
	if b.err != nil {
		return b
	}

	b.step++
	b.err = f()
	return b
}

func (b *Base) ReadBody() *Base {
	if b.err != nil {
		return b
	}

	b.step++
	defer b.r.Body.Close()
	b.reqBody, b.err = ioutil.ReadAll(b.r.Body)
	return b
}

func (b *Base) RequestJson(reqData interface{}) *Base {
	if b.err != nil {
		return b
	}

	b.step++
	b.err = json.Unmarshal(b.reqBody, reqData)
	return b
}

func (b *Base) ResponseJson(respData interface{}) *Base {
	if b.err != nil {
		return b
	}

	b.step++
	b.respBody, b.err = json.Marshal(respData)
	return b
}

func (b *Base) Write(w http.ResponseWriter) error {
	var err error
	if b.err == nil {
		b.step++
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(b.respBody)
		return err
	}

	switch reflect.TypeOf(b.err).String() {
	case "*json.SyntaxError":
		fallthrough
	case "*json.InvalidUTF8Error":
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte(b.err.Error()))
	default:
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte(b.err.Error()))
	}

	return err
}
