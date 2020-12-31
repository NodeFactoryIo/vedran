// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import models "github.com/NodeFactoryIo/vedran/internal/models"

// NodeRepository is an autogenerated mock type for the NodeRepository type
type NodeRepository struct {
	mock.Mock
}

// AddNodeToActive provides a mock function with given fields: ID
func (_m *NodeRepository) AddNodeToActive(ID string) error {
	ret := _m.Called(ID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(ID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FindByID provides a mock function with given fields: ID
func (_m *NodeRepository) FindByID(ID string) (*models.Node, error) {
	ret := _m.Called(ID)

	var r0 *models.Node
	if rf, ok := ret.Get(0).(func(string) *models.Node); ok {
		r0 = rf(ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Node)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetActiveNodes provides a mock function with given fields: selection
func (_m *NodeRepository) GetActiveNodes(selection string) *[]models.Node {
	ret := _m.Called(selection)

	var r0 *[]models.Node
	if rf, ok := ret.Get(0).(func(string) *[]models.Node); ok {
		r0 = rf(selection)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*[]models.Node)
		}
	}

	return r0
}

// GetAll provides a mock function with given fields:
func (_m *NodeRepository) GetAll() (*[]models.Node, error) {
	ret := _m.Called()

	var r0 *[]models.Node
	if rf, ok := ret.Get(0).(func() *[]models.Node); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*[]models.Node)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllActiveNodes provides a mock function with given fields:
func (_m *NodeRepository) GetAllActiveNodes() *[]models.Node {
	ret := _m.Called()

	var r0 *[]models.Node
	if rf, ok := ret.Get(0).(func() *[]models.Node); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*[]models.Node)
		}
	}

	return r0
}

// GetPenalizedNodes provides a mock function with given fields:
func (_m *NodeRepository) GetPenalizedNodes() (*[]models.Node, error) {
	ret := _m.Called()

	var r0 *[]models.Node
	if rf, ok := ret.Get(0).(func() *[]models.Node); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*[]models.Node)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IncreaseNodeCooldown provides a mock function with given fields: ID
func (_m *NodeRepository) IncreaseNodeCooldown(ID string) (*models.Node, error) {
	ret := _m.Called(ID)

	var r0 *models.Node
	if rf, ok := ret.Get(0).(func(string) *models.Node); ok {
		r0 = rf(ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Node)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsNodeActive provides a mock function with given fields: ID
func (_m *NodeRepository) IsNodeActive(ID string) bool {
	ret := _m.Called(ID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(ID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsNodeOnCooldown provides a mock function with given fields: ID
func (_m *NodeRepository) IsNodeOnCooldown(ID string) (bool, error) {
	ret := _m.Called(ID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(ID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveNodeFromActive provides a mock function with given fields: ID
func (_m *NodeRepository) RemoveNodeFromActive(ID string) error {
	ret := _m.Called(ID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(ID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ResetNodeCooldown provides a mock function with given fields: ID
func (_m *NodeRepository) ResetNodeCooldown(ID string) (*models.Node, error) {
	ret := _m.Called(ID)

	var r0 *models.Node
	if rf, ok := ret.Get(0).(func(string) *models.Node); ok {
		r0 = rf(ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Node)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: node
func (_m *NodeRepository) Save(node *models.Node) error {
	ret := _m.Called(node)

	var r0 error
	if rf, ok := ret.Get(0).(func(*models.Node) error); ok {
		r0 = rf(node)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateNodeUsed provides a mock function with given fields: node
func (_m *NodeRepository) UpdateNodeUsed(node models.Node) {
	_m.Called(node)
}
