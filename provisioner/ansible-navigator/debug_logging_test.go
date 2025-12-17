package ansiblenavigator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
