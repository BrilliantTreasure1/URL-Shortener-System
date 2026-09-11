package click

import (
	"encoding/json"
	"errors"

	clickDTO "analytics-service/application/click/dto"
	entities "analytics-service/entities"
	clickevent "analytics-service/repository/click-event"
)

var ErrInvalidPayload = errors.New("invalid click event payload")

type ConsumeClickEventUseCase struct {
	repo clickevent.ClickEventRepository
}

func NewConsumeClickEventUseCase(repo clickevent.ClickEventRepository) *ConsumeClickEventUseCase {
	return &ConsumeClickEventUseCase{
		repo: repo,
	}
}

func (c *ConsumeClickEventUseCase) Handle(payload []byte) error {

	var event clickDTO.LinkClickedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return ErrInvalidPayload
	}

	clickEvent, err := entities.NewClickEvent(
		event.EventID,
		event.UserID,
		event.ShortCode,
		event.OriginalURL,
		event.ClickedAt,
	)
	if err != nil {
		return ErrInvalidPayload
	}

	_, err = c.repo.Save(clickEvent)
	return err
}