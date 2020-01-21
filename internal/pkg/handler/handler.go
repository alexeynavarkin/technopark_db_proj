package handler

import (
	"encoding/json"
	"errors"
	"github.com/alexeynavarkin/technopark_db_proj/internal/pkg/api/view"
	"github.com/alexeynavarkin/technopark_db_proj/internal/pkg/consts"
	"github.com/alexeynavarkin/technopark_db_proj/internal/pkg/usecase"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"strconv"
	"strings"
)

type Handler struct {
	usecase *usecase.Usecase
	router  *fasthttprouter.Router
}

func NewHandler(usecase usecase.Usecase) Handler {
	h := Handler{
		usecase: &usecase,
		router:  fasthttprouter.New(),
	}

	h.router.POST("/api/user/:nickname/create", h.handleUserCreate)
	h.router.GET("/api/user/:nickname/profile", h.handleGetUserProfile)
	h.router.POST("/api/user/:nickname/profile", h.handleUserUpdate)

	h.router.POST("/api/forum/:slug/create", h.handleThreadCreate)
	h.router.GET("/api/forum/:slug/details", h.handleGetForumDetails)
	h.router.GET("/api/forum/:slug/threads", h.handleGetForumThreads)
	h.router.GET("/api/forum/:slug/users", h.handleGetForumUsers)

	h.router.POST("/api/thread/:slug_or_id/create", h.handlePostCreate)
	h.router.POST("/api/thread/:slug_or_id/vote", h.handleVoteForThread)
	h.router.GET("/api/thread/:slug_or_id/details", h.handleGetThreadDetails)
	h.router.POST("/api/thread/:slug_or_id/details", h.handleThreadUpdate)
	h.router.GET("/api/thread/:slug_or_id/posts", h.handleGetThreadPosts)

	h.router.GET("/api/post/:id/details", h.handleGetPostDetails)
	h.router.POST("/api/post/:id/details", h.handlePostUpdate)

	h.router.GET("/api/service/status", h.handleStatus)
	h.router.POST("/api/service/clear", h.handleClear)

	return h
}

func (h *Handler) GetHandleFunc() fasthttp.RequestHandler {
	return func(c *fasthttp.RequestCtx) {
		if string(c.Path()) == "/api/forum/create" {
			h.handleForumCreate(c)
		} else {
			h.router.Handler(c)
		}
	}
}

func (h *Handler) handleUserCreate(c *fasthttp.RequestCtx) {
	u := view.UserInput{}
	if err := json.Unmarshal(c.PostBody(), &u); err != nil {
		BadRequest(c, err)
		return
	}
	users, err := h.usecase.CreateUser(PathParam(c, "nickname"), u.Email, u.Fullname, u.About)
	if errors.Is(err, consts.ErrConflict) {
		Conflict(c, users)
		return
	}
	if err != nil {
		Error(c, err)
		return
	}
	Created(c, users[0])
}

func (h *Handler) handleGetUserProfile(c *fasthttp.RequestCtx) {
	u, err := h.usecase.GetUserByNickname(PathParam(c, "nickname"))
	if err != nil {
		Error(c, err)
		return
	}
	Ok(c, u)
}

func (h *Handler) handleUserUpdate(c *fasthttp.RequestCtx) {
	u := view.UserInput{}
	if err := json.Unmarshal(c.PostBody(), &u); err != nil {
		BadRequest(c, err)
		return
	}
	nick := PathParam(c, "nickname")
	user, err := h.usecase.UpdateUser(nick, u.Email, u.Fullname, u.About)
	if errors.Is(err, consts.ErrConflict) {
		ConflictWithMessage(c, err)
		return
	}
	if err != nil {
		Error(c, err)
		return
	}
	Ok(c, user)
}

func (h *Handler) handleForumCreate(c *fasthttp.RequestCtx) {
	forumToCreate := view.ForumCreate{}
	if err := json.Unmarshal(c.PostBody(), &forumToCreate); err != nil {
		BadRequest(c, err)
		return
	}
	forum, err := h.usecase.CreateForum(forumToCreate.Title, forumToCreate.Slug, forumToCreate.User)
	if errors.Is(err, consts.ErrConflict) {
		Conflict(c, forum)
		return
	}
	if err != nil {
		Error(c, err)
		return
	}
	Created(c, forum)
}

func (h *Handler) handleThreadCreate(c *fasthttp.RequestCtx) {
	thread := view.ThreadCreate{}
	if err := json.Unmarshal(c.PostBody(), &thread); err != nil {
		BadRequest(c, err)
		return
	}
	forum := PathParam(c, "slug")
	result, err := h.usecase.CreateThread(forum, thread)
	if errors.Is(err, consts.ErrConflict) {
		Conflict(c, result)
		return
	}
	if err != nil {
		Error(c, err)
		return
	}
	Created(c, result)
}

func (h *Handler) handleGetForumDetails(c *fasthttp.RequestCtx) {
	slug := PathParam(c, "slug")
	forum, err := h.usecase.GetForum(slug)
	if err != nil {
		Error(c, err)
		return
	}
	Ok(c, forum)
}

func (h *Handler) handleGetForumThreads(c *fasthttp.RequestCtx) {
	limit, _ := strconv.Atoi(QueryParam(c, "limit"))
	desc, _ := strconv.ParseBool(QueryParam(c, "desc"))
	threads, err := h.usecase.GetForumThreads(PathParam(c, "slug"), QueryParam(c, "since"), limit, desc)
	if err != nil {
		Error(c, err)
		return
	}
	Ok(c, threads)
}

func (h *Handler) handleGetForumUsers(c *fasthttp.RequestCtx) {
	limit, _ := strconv.Atoi(QueryParam(c, "limit"))
	desc, _ := strconv.ParseBool(QueryParam(c, "desc"))
	users, err := h.usecase.GetForumUsers(PathParam(c, "slug"), QueryParam(c, "since"), limit, desc)
	if err != nil {
		Error(c, err)
		return
	}
	Ok(c, users)
}

func (h *Handler) handlePostCreate(c *fasthttp.RequestCtx) {
	var posts []*view.PostCreate
	if err := json.Unmarshal(c.PostBody(), &posts); err != nil {
		BadRequest(c, err)
		return
	}
	result, err := h.usecase.CreatePosts(PathParam(c, "slug_or_id"), posts)
	if err != nil {
		Error(c, err)
		return
	}
	Created(c, result)
}

func (h *Handler) handleVoteForThread(c *fasthttp.RequestCtx) {
	var vote view.Vote
	if err := json.Unmarshal(c.PostBody(), &vote); err != nil {
		BadRequest(c, err)
		return
	}
	thread, err := h.usecase.VoteForThread(PathParam(c, "slug_or_id"), vote)
	if err != nil {
		Error(c, err)
		return
	}
	Ok(c, thread)
}

func (h *Handler) handleGetThreadDetails(c *fasthttp.RequestCtx) {
	thread, err := h.usecase.GetThread(PathParam(c, "slug_or_id"))
	if err != nil {
		Error(c, err)
		return
	}
	Ok(c, thread)
}

func (h *Handler) handleThreadUpdate(c *fasthttp.RequestCtx) {
	t := view.ThreadUpdate{}
	if err := json.Unmarshal(c.PostBody(), &t); err != nil {
		BadRequest(c, err)
		return
	}
	thread, err := h.usecase.UpdateThread(PathParam(c, "slug_or_id"), t.Message, t.Title)
	if err != nil {
		Error(c, err)
		return
	}
	Ok(c, thread)
}

func (h *Handler) handleGetThreadPosts(c *fasthttp.RequestCtx) {
	sp := QueryParam(c, "since")
	var since *int = nil
	if sp != "" {
		n, _ := strconv.Atoi(sp)
		since = &n
	}
	limit, _ := strconv.Atoi(QueryParam(c, "limit"))
	desc, _ := strconv.ParseBool(QueryParam(c, "desc"))
	posts, err := h.usecase.GetThreadPosts(
		PathParam(c, "slug_or_id"),
		limit,
		since,
		QueryParam(c, "sort"),
		desc,
	)
	if err != nil {
		Error(c, err)
		return
	}
	Ok(c, posts)
}

func (h *Handler) handleGetPostDetails(c *fasthttp.RequestCtx) {
	id, _ := strconv.Atoi(PathParam(c, "id"))
	related := strings.Split(QueryParam(c, "related"), ",")
	details, err := h.usecase.GetPostDetails(id, related)
	if err != nil {
		Error(c, err)
		return
	}
	result := map[string]interface{}{
		"post": details.Post,
	}
	for _, r := range related {
		switch r {
		case "user":
			result["author"] = details.Author
		case "forum":
			result["forum"] = details.Forum
		case "thread":
			result["thread"] = details.Thread
		}
	}
	Ok(c, result)
}

func (h *Handler) handlePostUpdate(c *fasthttp.RequestCtx) {
	t := view.PostUpdate{}
	if err := json.Unmarshal(c.PostBody(), &t); err != nil {
		BadRequest(c, err)
		return
	}
	id, _ := strconv.Atoi(PathParam(c, "id"))
	thread, err := h.usecase.UpdatePost(id, t.Message)
	if err != nil {
		Error(c, err)
		return
	}
	Ok(c, thread)
}

func (h *Handler) handleStatus(c *fasthttp.RequestCtx) {
	status, err := h.usecase.GetStatus()
	if err != nil {
		Error(c, err)
		return
	}
	Ok(c, status)
}

func (h *Handler) handleClear(c *fasthttp.RequestCtx) {
	err := h.usecase.Clear()
	if err != nil {
		Error(c, err)
		return
	}
	Ok(c, nil)
	return
}
