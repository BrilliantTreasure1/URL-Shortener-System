package link

import (
	"net/http"

	"url-shortener/application/link"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

type ResolveLinkController struct {
	linkUsecase link.ResolveShortLinkUseCase
}

func NewResolveLinkController(linkUsecase link.ResolveShortLinkUseCase) *ResolveLinkController {
	return &ResolveLinkController{
		linkUsecase: linkUsecase,
	}
}

func (rl *ResolveLinkController) Resolve(c *gin.Context) {

	ctx, span := otel.Tracer("resolve").Start(c.Request.Context(), "controller.resolve")
	defer span.End()

	shortCode := c.Param("short_code")

	link, err := rl.linkUsecase.ResolveShortLink(ctx, shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, link.OriginalURL())
}