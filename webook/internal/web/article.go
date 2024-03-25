package web

import "github.com/gin-gonic/gin"

var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
	//
}
func NewArticleHandler() *ArticleHandler{
	return &ArticleHandler{}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine){
	g := server.Group("/articles")
	g.POST("/edit", h.Edit)
}

func (h *ArticleHandler) Edit(ctx *gin.Context){
	//
}