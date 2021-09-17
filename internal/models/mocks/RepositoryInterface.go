// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	models "github.com/p7chkn/go-musthave-shortener-tpl/internal/models"
	mock "github.com/stretchr/testify/mock"
)

// RepositoryInterface is an autogenerated mock type for the RepositoryInterface type
type RepositoryInterface struct {
	mock.Mock
}

// AddURL provides a mock function with given fields: longURL, user
func (_m *RepositoryInterface) AddURL(longURL string, user string) string {
	ret := _m.Called(longURL, user)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(longURL, user)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetURL provides a mock function with given fields: shortURL
func (_m *RepositoryInterface) GetURL(shortURL string) (string, error) {
	ret := _m.Called(shortURL)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(shortURL)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(shortURL)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserURL provides a mock function with given fields: user
func (_m *RepositoryInterface) GetUserURL(user string) []models.ResponseGetURL {
	ret := _m.Called(user)

	var r0 []models.ResponseGetURL
	if rf, ok := ret.Get(0).(func(string) []models.ResponseGetURL); ok {
		r0 = rf(user)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.ResponseGetURL)
		}
	}

	return r0
}
