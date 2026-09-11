package click

import (
	entities "analytics-service/entities"
	clickevent "analytics-service/repository/click-event"
)

type mockClickRepo struct {
	savedEvent *entities.ClickEvent
	saveResult bool
	saveErr    error
	calledSave bool
}

func (m *mockClickRepo) Save(clickEvent *entities.ClickEvent) (bool, error) {
	m.calledSave = true
	m.savedEvent = clickEvent
	return m.saveResult, m.saveErr
}

var _ clickevent.ClickEventRepository = (*mockClickRepo)(nil)