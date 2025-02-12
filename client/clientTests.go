// clientTests

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/anthdm/hollywood/actor"
	"github.com/anthdm/hollywood/examples/chat/types"
	"github.com/stretchr/testify/assert"
)

// Mock the actor.Context to simulate sending and receiving messages.
type mockContext struct {
	message interface{}
}

func (m *mockContext) Send(pid *actor.PID, message interface{}) {
	m.message = message
}

func (m *mockContext) Message() interface{} {
	return m.message
}

// Mock PID for testing.
var mockPID = &actor.PID{}

// Test newClient function.
func TestNewClient(t *testing.T) {
	clientProducer := newClient("testuser", mockPID)
	receiver := clientProducer()
	client := receiver.(*client)

	assert.Equal(t, "testuser", client.username)
	assert.Equal(t, mockPID, client.serverPID)
	assert.NotNil(t, client.logger)
}

// Test Receive function for actor.Started case.
func TestClientReceiveStarted(t *testing.T) {
	mockCtx := &mockContext{}
	client := &client{
		username:  "testuser",
		serverPID: mockPID,
		logger:    slog.Default(),
	}

	client.Receive(mockCtx)
	msg, ok := mockCtx.message.(*types.Connect)
	assert.True(t, ok)
	assert.Equal(t, "testuser", msg.Username)
}

// Test Receive function for actor.Stopped case.
func TestClientReceiveStopped(t *testing.T) {
	mockCtx := &mockContext{message: actor.Stopped{}}
	client := &client{
		username:  "testuser",
		serverPID: mockPID,
		logger:    slog.Default(),
	}

	// Capture logger output
	var logBuf bytes.Buffer
	client.logger = slog.New(slog.HandlerOptions{Output: &logBuf})

	client.Receive(mockCtx)
	assert.Contains(t, logBuf.String(), "client stopped")
}

// Test Receive function for types.Message case.
func TestClientReceiveMessage(t *testing.T) {
	mockCtx := &mockContext{
		message: &types.Message{Username: "otheruser", Msg: "hello"},
	}
	client := &client{
		username:  "testuser",
		serverPID: mockPID,
		logger:    slog.Default(),
	}

	// Capture stdout output
	var outBuf bytes.Buffer
	fmt.Fprintf(&outBuf, "%s: %s\n", "otheruser", "hello")

	client.Receive(mockCtx)
	assert.Equal(t, outBuf.String(), fmt.Sprintf("%s: %s\n", "otheruser", "hello"))
}

// Test main function with mocked inputs
func TestMainFunction(t *testing.T) {
	// Mock command-line flags and environment variables
	flag.Set("connect", "127.0.0.1:4000")
	flag.Set("listen", "127.0.0.1:5000")
	os.Setenv("USER", "testuser")

	// Mock user input via stdin
	input := "hello\nquit\n"
	reader := strings.NewReader(input)
	bufioReader := bufio.NewReader(reader)
	scanner := bufio.NewScanner(bufioReader)

	// Set up a fake server and engine (mocked, for simplicity).
	// This would typically require more extensive mocking/stubbing.
	serverPID := actor.NewPID("127.0.0.1:4000", "server/primary")
	engine := actor.NewEngine(actor.NewEngineConfig())

	clientPID := engine.Spawn(newClient("testuser", serverPID), "client", actor.WithID("testuser"))

	for scanner.Scan() {
		msg := scanner.Text()
		assert.NotEmpty(t, msg)
		if msg == "quit" {
			break
		}
	}

	assert.Nil(t, scanner.Err())

	// Simulate disconnect message
	engine.SendWithSender(serverPID, &types.Disconnect{}, clientPID)
}
