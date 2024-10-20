package forwarder

import (
	"testing"
	"viktig/internal/entities"

	"github.com/stretchr/testify/assert"
)

func TestRender(t *testing.T) {
	t.Run("new message", func(t *testing.T) {
		message := entities.Message{
			Type:       entities.MessageTypeNew,
			Text:       "Hello",
			VkSenderId: 1234,
		}
		actual := render(message)
		expected := "👤 <a href=\"https://vk.com/id1234\">1234</a>\n💬 Hello"
		assert.Equal(t, expected, actual)
	})
	t.Run("edited message", func(t *testing.T) {
		message := entities.Message{
			Type:       entities.MessageTypeEdit,
			Text:       "Edit",
			VkSenderId: 1234,
		}
		actual := render(message)
		expected := "👤 <a href=\"https://vk.com/id1234\">1234</a>\n✏️ Edit"
		assert.Equal(t, expected, actual)
	})
	t.Run("edited by community message", func(t *testing.T) {
		message := entities.Message{
			Type:       entities.MessageTypeEdit,
			Text:       "Edit",
			VkSenderId: -123,
		}
		actual := render(message)
		expected := "👤 <a href=\"https://vk.com/club123\">123</a>\n✏️ Edit"
		assert.Equal(t, expected, actual)
	})
	t.Run("replied message", func(t *testing.T) {
		message := entities.Message{
			Type:       entities.MessageTypeReply,
			Text:       "Reply",
			VkSenderId: 4321,
		}
		actual := render(message)
		expected := "👤 <a href=\"https://vk.com/id4321\">4321</a>\n↩️ Reply"
		assert.Equal(t, expected, actual)
	})
	t.Run("with sender name", func(t *testing.T) {
		message := entities.Message{
			Type:       entities.MessageTypeNew,
			Text:       "Hello",
			VkSenderId: 1234,
			VkSender:   &entities.VkUser{FirstName: "John", LastName: "Doe"},
		}
		actual := render(message)
		expected := "👤 <a href=\"https://vk.com/id1234\">John Doe</a>\n💬 Hello"
		assert.Equal(t, expected, actual)
	})
	t.Run("escape HTML", func(t *testing.T) {
		message := entities.Message{
			Type:       entities.MessageTypeNew,
			Text:       "<a href=\"https://x.com\">&</a>",
			VkSenderId: 1,
		}
		actual := render(message)
		expected := "👤 <a href=\"https://vk.com/id1\">1</a>\n💬 &lt;a href=&#34;https://x.com&#34;&gt;&amp;&lt;/a&gt;"
		assert.Equal(t, expected, actual)
	})
}
