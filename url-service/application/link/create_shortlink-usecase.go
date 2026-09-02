package link

import (
	"errors"

	entities "url-shortener/entities/link"
	linkRepo "url-shortener/repository/link"
)

type CreateShortLinkUseCase struct {
	shortCodeGenerator ShortCodeGenerator
	linkRepo linkRepo.LinkRepository
}

func NewCreateShortLinkUseCase (

	shortCodeGenerator ShortCodeGenerator ,
	linkRepo linkRepo.LinkRepository, 

	) *CreateShortLinkUseCase {
	return &CreateShortLinkUseCase{
		shortCodeGenerator: shortCodeGenerator,
		linkRepo: linkRepo,
	}
}

func (c *CreateShortLinkUseCase) CreateShortLink (
	userID int,
	originalURL string,
) (*entities.Link , error) {

	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}

	if originalURL == "" {
		return nil, errors.New("original url cannot be empty")
	}

	codeValue, err := c.linkRepo.NextCodeValue()
	if err != nil {
		return nil, err
	}

	shortCode, err := c.shortCodeGenerator.GenerateUnique(codeValue, 5)
	if err != nil {
		return nil, err
	}

	link, err := entities.NewLink(nil, userID, originalURL, shortCode)
	if err != nil {
		return nil, err
	}

	createdLink, err := c.linkRepo.Create(link)
	if err != nil {
		return nil, err
	}

	return createdLink, nil
}



