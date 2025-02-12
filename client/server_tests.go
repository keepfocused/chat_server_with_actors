// server_tests
package main

import (
	"testing"

	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/hollywood/examples/chat/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockContext is a mock for actor.Context
type MockContext struct {
	mock.Mock
	sender *actor.PID
	msg    interface{}
}

func (m *MockContext) Message() interface{} {
	return m.msg
}

func (m *MockContext) Sender() *actor.PID {
	return m.sender
}

func (m *MockContext) Forward(pid *actor.PID) {
	m.Called(pid)
}

func (m *MockContext) Self() *actor.PID {
	return nil
}

// MockPID is a mock for actor.PID
type MockPID struct {
	ID      string
	Address string
}

func (pid *MockPID) Equals(other *actor.PID) bool {
	return pid.ID == other.ID && pid.Address == other.Address
}

func TestServerHandleConnect(t *testing.T) {
	server := newServer().(*server)

	// Mock Context and PID
	ctx := &MockContext{}
	pid := &MockPID{ID: "1", Address: "client1"}
	ctx.sender = &actor.PID{ID: pid.ID, Address: pid.Address}
	ctx.msg = &types.Connect{Username: "user1"}

	// Simulate Connect message
	server.Receive(ctx)

	// Check that the client and user were added
	assert.Equal(t, server.clients[pid.Address], ctx.sender)
	assert.Equal(t, server.users[pid.Address], "user1")
}

func TestServerHandleDisconnect(t *testing.T) {
	server := newServer().(*server)

	// Mock Context and PID
	ctx := &MockContext{}
	pid := &MockPID{ID: "1", Address: "client1"}
	server.clients[pid.Address] = &actor.PID{ID: pid.ID, Address: pid.Address}
	server.users[pid.Address] = "user1"

	// Simulate Disconnect message
	ctx.sender = &actor.PID{ID: pid.ID, Address: pid.Address}
	ctx.msg = &types.Disconnect{}
	server.Receive(ctx)

	// Check that the client and user were removed
	_, clientExists := server.clients[pid.Address]
	_, userExists := server.users[pid.Address]
	assert.False(t, clientExists)
	assert.False(t, userExists)
}

func TestServerHandleMessage(t *testing.T) {
	server := newServer().(*server)

	// Mock Context and PIDs
	ctx := &MockContext{}
	pid1 := &MockPID{ID: "1", Address: "client1"}
	pid2 := &MockPID{ID: "2", Address: "client2"}
	ctx.sender = &actor.PID{ID: pid1.ID, Address: pid1.Address}
	server.clients[pid1.Address] = &actor.PID{ID: pid1.ID, Address: pid1.Address}
	server.clients[pid2.Address] = &actor.PID{ID: pid2.ID, Address: pid2.Address}

	// Set message
	ctx.msg = &types.Message{Msg: "Hello"}

	// Mock Forward behavior
	ctx.On("Forward", mock.Anything).Return(nil)

	// Simulate Message handling
	server.Receive(ctx)

	// Verify that message was forwarded to pid2 (and not pid1)
	ctx.AssertCalled(t, "Forward", server.clients[pid2.Address])
	ctx.AssertNotCalled(t, "Forward", server.clients[pid1.Address])
}
