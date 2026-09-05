package container

import (

	"database/sql"

	applicationLink "url-shortener/application/link"
	controllerLink "url-shortener/controller/link"
	repositoryLink "url-shortener/repository/link"
	"url-shortener/service/shortcode-generator"
)

type LinkContainer struct {
	CreateLinkController  *controllerLink.CreateLinkController
	ResolveLinkController *controllerLink.ResolveLinkController
	ListLinkController    *controllerLink.ListLinkController
}

func NewLinkContainer(db *sql.DB) (*LinkContainer, error) {

	linkRepository := repositoryLink.NewLinkRepositoryPostgresql(db)
	shortcodeGenerator := shortcodegenerator.NewUniqueGenerator()

	createShortLinkUseCase := applicationLink.NewCreateShortLinkUseCase(
		shortcodeGenerator,
		linkRepository,
	)

	createLinkController := controllerLink.NewCreateLinkController(
		*createShortLinkUseCase,
	)

	resolveShortLinkUseCase := applicationLink.NewResolveShortLinkUseCase(
		linkRepository,
	)

	resolveLinkController := controllerLink.NewResolveLinkController(
		*resolveShortLinkUseCase,
	)

	listShortLinkUseCase := applicationLink.NewListShortLinkUseCase(
		linkRepository,
	)

	listLinkController := controllerLink.NewListLinkController(
		*listShortLinkUseCase,
	)

	return &LinkContainer{
		CreateLinkController:  createLinkController,
		ResolveLinkController: resolveLinkController,
		ListLinkController:    listLinkController,
	}, nil
}