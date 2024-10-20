# VikTig

Service for forwarding messages from VK communities to Telegram chats.

## Running

1. Create a YAML configuration file.
    ```yaml
    # Create a bot and get the token with https://t.me/BotFather
    tg_bot_token: 1234567890:xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    # Get a VK access token of any type
    # https://dev.vk.com/en/api/access-token/getting-started
    vk_api_token: vk1.a.xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

    # List of VK communities
    communities:
    - hook_id: my-community  # Used in VK callback URL: /api/vk/callback/<hook_id>
      confirmation_string: abcde123  # From VK community Callback API settings
      tg_chat_id: 123456789  # Find your ID with https://t.me/userinfobot
    ```
1. Run the service
    ```shell
    go run cmd/app/main.go --config my-config.yml
    ```
