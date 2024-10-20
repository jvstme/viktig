package telegram_bot_handler

import (
	"fmt"
	"log/slog"
	"strconv"
	"viktig/internal/entities"
	"viktig/internal/repository"
	"viktig/internal/services/http_server/handlers"

	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	tele "gopkg.in/telebot.v3"
)

// todo: use real public host instead of localhost
var interactionSavedMsg = `Your community is ready! Configure the community's settings on VK.com to start receiving messages:

1. Return to your community's Callback API settings
2. Set API version: 5.199
3. Set URL: https://localhost/api/vk/callback/%s
4. Click Confirm
5. Go to the Event types tab
6. Check Message received, Message sent, and Message edited
`

type telegramBotHandler struct {
	bot    *tele.Bot
	repo   repository.Repository
	logger *slog.Logger
}

func New(bot *tele.Bot, repo repository.Repository, logger *slog.Logger) handlers.Handler {
	h := &telegramBotHandler{bot: bot, repo: repo, logger: logger.With("name", "TelegramBotHandler")}
	h.bot.Handle("/start", h.enterMainMenu)
	h.bot.Handle("/add", h.handleAdd)
	h.bot.Handle(tele.OnText, h.handleText)
	return h
}

func (h *telegramBotHandler) Handle(ctx *fasthttp.RequestCtx) {
	update := tele.Update{}
	if err := jsoniter.Unmarshal(ctx.Request.Body(), &update); err != nil {
		h.logger.Error(fmt.Sprintf("error handling request: %+v", err))
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	h.logger.Info("received Telegram update", "id", update.ID)
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	h.bot.ProcessUpdate(update)
	// todo: store users in db
}

func (h *telegramBotHandler) enterMainMenu(ctx tele.Context) error {
	interactions, err := h.repo.ListInteractions(int(ctx.Sender().ID))
	if err != nil {
		h.logger.Error("error listing interactions", "err", err.Error())
		return ctx.Send("Could not list your communities. Please try again")
	}
	return ctx.Send(renderMainMenu(interactions))
}

func (h *telegramBotHandler) handleAdd(ctx tele.Context) error {
	interaction := &entities.IncompleteInteraction{
		UserId: int(ctx.Sender().ID),
	}
	if err := h.repo.StoreIncompleteInteraction(interaction); err != nil {
		return err
	}
	return ctx.Send("Enter VK community name to forward messages from")
}

func (h *telegramBotHandler) handleText(ctx tele.Context) error {
	interaction, err := h.repo.GetIncompleteInteraction(int(ctx.Sender().ID))
	if err != nil {
		return h.enterMainMenu(ctx)
	}

	// todo: delete IncompleteInteraction if user cancels editing

	if interaction.Name == nil {
		return h.handleInteractionName(interaction, ctx)
	} else if interaction.TgChatId == nil {
		return h.handleInteractionChatId(interaction, ctx)
	} else if interaction.ConfirmationString == nil {
		return h.handleInteractionConfirmationString(interaction, ctx)
	}

	return h.enterMainMenu(ctx)
}

func (h *telegramBotHandler) handleInteractionName(interaction *entities.IncompleteInteraction, ctx tele.Context) error {
	interaction.Name = &ctx.Message().Text
	if err := h.repo.UpdateIncompleteInteraction(interaction); err != nil {
		h.logger.Error("error saving interaction name", "err", err.Error())
		return ctx.Send("Could not save the name. Please try again")
	}
	return ctx.Send(fmt.Sprintf(
		"Enter Telegram chat ID to forward messages to. "+
			"In case you want to forward messages to yourself, your ID is %v",
		ctx.Sender().ID,
	))
}

func (h *telegramBotHandler) handleInteractionChatId(interaction *entities.IncompleteInteraction, ctx tele.Context) error {
	chatId, err := strconv.Atoi(ctx.Message().Text)
	if err != nil {
		return ctx.Send("Telegram chat ID should be an integer. Please try again")
	}
	// todo: check that user is authorized to configure this chat id
	interaction.TgChatId = &chatId
	if err := h.repo.UpdateIncompleteInteraction(interaction); err != nil {
		h.logger.Error("error saving interaction chat id", "err", err.Error())
		return ctx.Send("Could not save the chat ID. Please try again")
	}
	return ctx.Send(
		"Enter VK community's confirmation string. Find it on VK.com in your " +
			"community settings → API usage → Callback API → String to be returned.\n" +
			"Don't close the settings page, you will need it in a moment",
	)
}

func (h *telegramBotHandler) handleInteractionConfirmationString(interaction *entities.IncompleteInteraction, ctx tele.Context) error {
	interaction.ConfirmationString = &ctx.Message().Text
	completeInteraction, err := finalizeInteraction(h.repo, interaction)
	if err != nil {
		h.logger.Error("error finalizing interaction", "err", err.Error())
		return ctx.Send("Could not save the community. Please try again")
	}
	h.logger.Info("saved new interaction", "id", completeInteraction.Id)
	return ctx.Send(fmt.Sprintf(interactionSavedMsg, completeInteraction.Id))
}

func finalizeInteraction(
	repo repository.Repository, interaction *entities.IncompleteInteraction,
) (*entities.Interaction, error) {
	if err := repo.DeleteIncompleteInteraction(interaction.UserId); err != nil {
		return nil, err
	}
	completeInteraction := &entities.Interaction{
		Id:                 uuid.New(),
		UserId:             interaction.UserId,
		Name:               *interaction.Name,
		ConfirmationString: *interaction.ConfirmationString,
		TgChatId:           *interaction.TgChatId,
	}
	if err := repo.StoreInteraction(completeInteraction); err != nil {
		return nil, err
	}
	return completeInteraction, nil
}

func renderMainMenu(interactions []*entities.Interaction) string {
	result := "Welcome to viktig!"
	if len(interactions) > 0 {
		result += " These are your connected communities:\n\n"
		for i, interaction := range interactions {
			result += fmt.Sprintf("%d. %s\n", i+1, interaction.Name)
		}
	} else {
		result += "\n"
	}
	result += "\n/add a new community"
	return result
}
