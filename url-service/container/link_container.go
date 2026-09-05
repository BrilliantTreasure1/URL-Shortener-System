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

	return &LinkContainer{
		CreateLinkController:  createLinkController,
		ResolveLinkController: resolveLinkController,
	}, nil
}