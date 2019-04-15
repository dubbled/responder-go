package respond

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type loginModel struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func checkResponder(t *testing.T, b *Base) {
	if b.err != nil {
		t.Error(b.err)
	}
}

func TestCustomFn(t *testing.T) {
	r := httptest.NewRequest("GET", "http://example.com", nil)
	var tVal string

	responder := &Base{r: r}
	responder.
		CustomFn(func() error {
			tVal = "ok"
			return nil
		})

	assert.Equal(t, responder.err, nil)
	assert.Equal(t, responder.step, 1)
	assert.Equal(t, tVal, "ok")

	responder = &Base{r: r}
	responder.
		CustomFn(func() error {
			tVal = "not ok"
			return errors.New("failure")
		})

	assert.NotEqual(t, responder.err, nil)
	assert.Equal(t, responder.step, 1)
}

func TestJsonSuccess(t *testing.T) {
	body := []byte(`
	{
		"username":"testing",
		"password":"password"
	}`,
	)

	r := httptest.NewRequest("GET", "http://example.com", bytes.NewReader(body))
	responder := &Base{r: r}
	var token string
	mockWriter := httptest.NewRecorder()
	login := &loginModel{}

	err := responder.
		ReadBody().
		RequestJson(login).
		CustomFn(func() error { // imaginary db op that obtains a token
			token = "session-token"
			return nil
		}).
		ResponseJson(token).
		Write(mockWriter)

	resp := &bytes.Buffer{}
	resp.ReadFrom(mockWriter.Result().Body)

	assert.Equal(t, responder.err, nil)
	assert.Equal(t, responder.step, 5)
	assert.Equal(t, login.Username, "testing")
	assert.Equal(t, login.Password, "password")
	assert.Equal(t, err, nil)
	assert.Equal(t, mockWriter.Result().StatusCode, 200)
	assert.Equal(t, resp.String(), "\""+token+"\"")
}

func TestJsonFailure(t *testing.T) {
	body := []byte(`
	{
		"username":"invalid"
		"password":json
	}`,
	)

	r := httptest.NewRequest("GET", "http://example.com", bytes.NewReader(body))
	responder := &Base{r: r}
	login := &loginModel{}
	mockWriter := httptest.NewRecorder()

	err := responder.
		ReadBody().
		RequestJson(login).
		ResponseJson([]string{"never", "gets", "here"}).
		Write(mockWriter)

	resp := &bytes.Buffer{}
	resp.ReadFrom(mockWriter.Result().Body)

	assert.Equal(t, err, nil)
	assert.NotEqual(t, responder.err, nil)
	assert.Equal(t, responder.step, 2)
	assert.Equal(t, err, nil)
	assert.Equal(t, mockWriter.Result().StatusCode, http.StatusBadRequest)
}
