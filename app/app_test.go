package app

import (
	"awesome/todos/model"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTodos(t *testing.T) {
	// 테스트를 위해서 세션을 가져오는 함수를 재정의한다.
	getSessionID = func(r *http.Request) string {
		return "testSessionId"
	}
	os.Remove("./test.db")
	assert := assert.New(t)

	appHandler := MakeHandler("./test.db")
	defer appHandler.Close()

	ts := httptest.NewServer(appHandler.Handler)
	defer ts.Close()

	// add에서 FormValue로 받았기 때문에 PostForm을 사용하여야 함.
	resp, err := http.PostForm(ts.URL+"/todos", url.Values{"name": {"Test todo"}})
	assert.NoError(err)
	assert.Equal(http.StatusCreated, resp.StatusCode)

	todo := new(model.Todo)
	err = json.NewDecoder(resp.Body).Decode(&todo)
	assert.NoError(err)
	assert.Equal(todo.Name, "Test todo")
	id1 := todo.ID

	// todo2를 하나더 추가한다.
	resp, err = http.PostForm(ts.URL+"/todos", url.Values{"name": {"Test todo2"}})
	assert.NoError(err)
	assert.Equal(http.StatusCreated, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&todo)
	assert.NoError(err)
	assert.Equal(todo.Name, "Test todo2")
	id2 := todo.ID

	// todo 목록을 받아와서 추가한 데이터가 맞는지 테스트한다.
	resp, err = http.Get(ts.URL + "/todos")
	assert.NoError(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	todos := []*model.Todo{}
	err = json.NewDecoder(resp.Body).Decode(&todos)
	assert.NoError(err)
	assert.Equal(2, len(todos))
	for _, t := range todos {
		if t.ID == id1 {
			assert.Equal(t.Name, "Test todo")
		} else if t.ID == id2 {
			assert.Equal(t.Name, "Test todo2")
		} else {
			assert.Error(fmt.Errorf("testID should be id1 or id2, received id is " + strconv.Itoa(t.ID)))
		}
	}

	// 완료 처리를 테스트한다.
	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/todos/"+strconv.Itoa(id1), strings.NewReader("{\"completed\":true}"))
	assert.NoError(err)
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(err)
	assert.Equal(resp.StatusCode, http.StatusOK)

	// 완료처리로 데이터가 바뀌었는지 확인한다.
	resp, err = http.Get(ts.URL + "/todos")
	assert.NoError(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	todos = []*model.Todo{}
	err = json.NewDecoder(resp.Body).Decode(&todos)
	assert.NoError(err)
	assert.Equal(2, len(todos))
	for _, t := range todos {
		if t.ID == id1 {
			assert.Equal(t.Completed, true)
		}
	}

	// 삭제 테스트를 한다.
	req, _ = http.NewRequest(http.MethodDelete, ts.URL+"/todos/"+strconv.Itoa(id1), nil)
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(err)
	assert.Equal(resp.StatusCode, http.StatusOK)

	resp, err = http.Get(ts.URL + "/todos")
	assert.NoError(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	todos = []*model.Todo{}
	err = json.NewDecoder(resp.Body).Decode(&todos)
	assert.NoError(err)
	assert.Equal(1, len(todos))
	for _, t := range todos {
		if t.ID == id1 {
			assert.Error(fmt.Errorf("id1 is exists"))
		}
	}
}
