package handlers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
}

func (mockRepo) Migrate(data interface{}) error {
	panic("implement me")
}

func (m mockRepo) Put(data interface{}) error {
	ret := m.Called(data)
	return ret.Error(0)
}

func (mockRepo) Del(cond interface{}) error {
	panic("implement me")
}

func (mockRepo) GetOne(cond interface{}, data interface{}) error {
	panic("implement me")
}

func (mockRepo) GetAll(cond interface{}, data interface{}) error {
	panic("implement me")
}

func (mockRepo) Close() error {
	panic("implement me")
}

var (
	_mock = &mockRepo{}
)

func TestLunchAi_Add(t *testing.T) {
	ai, _ := newLunchStore()
	ai.repo = _mock

	r := Restaurant{Name: "restaurant1"}
	err := errors.New("err1")
	_call := _mock.On("Put", r)

	_call.Return(nil)
	assert.Equal(t, ai.Add(r), nil)

	_call.Return(err)
	assert.Equal(t, ai.Add(r), err)
}
