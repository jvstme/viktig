package forwarder

import (
	"testing"
	"viktig/internal/core"

	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	actual := render(core.Message{Text: "Hello", VkSenderId: 1234})
	expected := "👤 <a href=\"https://vk.com/id1234\">1234</a>\n💬 Hello"
	assert.Equal(t, expected, actual)
}

func TestRenderEscapesHTML(t *testing.T) {
	actual := render(core.Message{Text: "<a href=\"https://x.com\">&</a>", VkSenderId: 1})
	expected := "👤 <a href=\"https://vk.com/id1\">1</a>\n💬 &lt;a href=&#34;https://x.com&#34;&gt;&amp;&lt;/a&gt;"
	assert.Equal(t, expected, actual)
}
