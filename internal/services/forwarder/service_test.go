package forwarder

import (
	"testing"
	"viktig/internal/entities"

	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	t.Run("basic render", func(t *testing.T) {
		actual := render(entities.Message{Text: "Hello", VkSenderId: 1234})
		expected := "👤 <a href=\"https://vk.com/id1234\">1234</a>\n💬 Hello"
		assert.Equal(t, expected, actual)
	})
	t.Run("escape HTML", func(t *testing.T) {
		actual := render(entities.Message{Text: "<a href=\"https://x.com\">&</a>", VkSenderId: 1})
		expected := "👤 <a href=\"https://vk.com/id1\">1</a>\n💬 &lt;a href=&#34;https://x.com&#34;&gt;&amp;&lt;/a&gt;"
		assert.Equal(t, expected, actual)
	})
}
