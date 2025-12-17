package ansiblenavigatorlocal

import (
	"fmt"
	"io"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/require"
)

type mockUi struct {
	messageMessages []string
}

func (m *mockUi) Ask(string) (string, error)                  { return "", nil }
func (m *mockUi) Askf(string, ...interface{}) (string, error) { return "", nil }
func (m *mockUi) Say(string)                                  {}
func (m *mockUi) Sayf(string, ...interface{})                 {}
func (m *mockUi) Message(message string)                      { m.messageMessages = append(m.messageMessages, message) }
func (m *mockUi) Messagef(message string, args ...interface{}) {
	m.messageMessages = append(m.messageMessages, fmt.Sprintf(message, args...))
}
func (m *mockUi) Error(string)                  {}
func (m *mockUi) Errorf(string, ...interface{}) {}
func (m *mockUi) Machine(string, ...string)     {}
func (m *mockUi) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	return stream
}

func newMockUi() packersdk.Ui { return &mockUi{messageMessages: make([]string, 0)} }

func TestIsPluginDebugEnabled(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		require.False(t, isPluginDebugEnabled(nil))
	})

	t.Run("missing logging", func(t *testing.T) {
		require.False(t, isPluginDebugEnabled(&NavigatorConfig{}))
	})

	t.Run("non-debug", func(t *testing.T) {
		require.False(t, isPluginDebugEnabled(&NavigatorConfig{Logging: &LoggingConfig{Level: "info"}}))
	})

	t.Run("debug case-insensitive", func(t *testing.T) {
		require.True(t, isPluginDebugEnabled(&NavigatorConfig{Logging: &LoggingConfig{Level: "DeBuG"}}))
	})
}

func TestDebugf_GatedAndPrefixed(t *testing.T) {
	t.Parallel()

	t.Run("disabled", func(t *testing.T) {
		ui := newMockUi().(*mockUi)
		debugf(ui, false, "hello")
		require.Empty(t, ui.messageMessages)
	})

	t.Run("enabled", func(t *testing.T) {
		ui := newMockUi().(*mockUi)
		debugf(ui, true, "hello %s", "world")
		require.Len(t, ui.messageMessages, 1)
		require.Equal(t, "[DEBUG] hello world", ui.messageMessages[0])
	})
}
