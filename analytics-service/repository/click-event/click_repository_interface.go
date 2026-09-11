package clickevent

import entities "analytics-service/entities"

type ClickEventRepository interface {
	Save(clickEvent *entities.ClickEvent) (bool, error)
}