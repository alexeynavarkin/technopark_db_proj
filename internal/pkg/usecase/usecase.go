package usecase

import (
	"fmt"
	"github.com/alexeynavarkin/technopark_db_proj/internal/pkg/api/view"
	"github.com/alexeynavarkin/technopark_db_proj/internal/pkg/consts"
	"github.com/alexeynavarkin/technopark_db_proj/internal/pkg/model"
	"github.com/alexeynavarkin/technopark_db_proj/internal/pkg/repository"
	"time"
)

type Usecase struct {
	repo repository.Repository
}

func NewUsecase(repo repository.Repository) Usecase {
	return Usecase{repo: repo}
}

func (u *Usecase) GetUserByNickname(nickname string) (*model.User, error) {
	return u.repo.GetUserByNickname(nickname)
}

func (u *Usecase) CreateUser(nickname, email, fullname, about string) ([]*model.User, error) {
	existing, err := u.repo.GetUsersByNicknameOrEmail(nickname, email)
	if err != nil && err != consts.ErrNotFound {
		return nil, err
	}
	if existing != nil {
		return existing, consts.ErrConflict
	}
	user, err := u.repo.CreateUser(nickname, email, fullname, about)
	return []*model.User{user}, err
}

func (u *Usecase) UpdateUser(nickname, email, fullname, about string) (*model.User, error) {
	userToUpdate, err := u.repo.GetUserByNickname(nickname)
	if err != nil {
		return nil, err
	}
	if email == "" {
		email = userToUpdate.Email
	}
	if fullname == "" {
		fullname = userToUpdate.Fullname
	}
	if about == "" {
		about = userToUpdate.About
	}
	if err := u.repo.UpdateUserByNickname(nickname, email, fullname, about); err != nil {
		return nil, err
	}
	return u.repo.GetUserByNickname(nickname)
}

func (u *Usecase) CreateForum(title, slug, nickname string) (*model.Forum, error) {
	userNick, err := u.repo.GetUserNickname(nickname)
	if err != nil {
		return nil, err
	}

	existingForum, err := u.repo.GetForumBySlug(slug)
	if err != nil && err != consts.ErrNotFound {
		return nil, err
	}
	if existingForum != nil {
		return existingForum, fmt.Errorf("%w: forum with this slug already exists", consts.ErrConflict)
	}

	return u.repo.CreateForum(title, slug, userNick)
}

func (u *Usecase) CreateThread(forumSlug string, thread view.ThreadCreate) (*model.Thread, error) {
	if _, err := u.repo.GetUserNickname(thread.Author); err != nil {
		return nil, err
	}
	forum, err := u.repo.GetForumSlug(forumSlug)
	if err != nil {
		return nil, err
	}

	if thread.Slug != "" {
		existing, err := u.repo.GetThreadBySlug(thread.Slug)
		if err != nil && err != consts.ErrNotFound {
			return nil, err
		}
		if existing != nil {
			return existing, fmt.Errorf("%w: thread with this slug already exists", consts.ErrConflict)
		}
	}

	if thread.Created == "" {
		thread.Created = time.Now().Format(time.RFC3339)
	}

	return u.repo.CreateThread(forum, thread)
}

func (u *Usecase) UpdateThread(threadSlugOrID string, message, title string) (*model.Thread, error) {
	return u.repo.UpdateThread(threadSlugOrID, message, title)
}

func (u *Usecase) CreatePosts(threadSlugOrID string, posts []*view.PostCreate) (model.Posts, error) {
	thread, err := u.repo.GetThreadFieldsBySlugOrID("id, forum", threadSlugOrID)
	if err != nil {
		return nil, err
	}
	if err := u.CheckPostsCreate(posts, thread.ID); err != nil {
		return nil, err
	}
	return u.repo.CreatePosts(posts, thread)
}

func (u *Usecase) CheckPostsCreate(posts []*view.PostCreate, threadID int) error {
	for _, post := range posts {
		if err := u.CheckPostCreate(post, threadID); err != nil {
			return err
		}
	}
	return nil
}

func (u *Usecase) CheckPostCreate(post *view.PostCreate, threadID int) error {
	if _, err := u.repo.GetUserNickname(post.Author); err != nil {
		return err
	}
	if post.Parent != 0 {
		parent, err := u.repo.GetPostByID(post.Parent)
		if err == consts.ErrNotFound {
			return fmt.Errorf("%w: post parent do not exists", consts.ErrConflict)
		}
		if err != nil {
			return err
		}
		if parent.Thread != threadID {
			return fmt.Errorf("%w: parent post was created in another thread", consts.ErrConflict)
		}
	}
	return nil
}

func (u *Usecase) GetForum(slug string) (*model.Forum, error) {
	return u.repo.GetForumBySlug(slug)
}

func (u *Usecase) GetForumThreads(forumSlug, since string, limit int, desc bool) (model.Threads, error) {
	forum, err := u.repo.GetForumSlug(forumSlug)
	if err != nil {
		return nil, err
	}
	var threads model.Threads
	if since == "" {
		threads, err = u.repo.GetForumThreads(forum.Slug, limit, desc)
	} else {
		threads, err = u.repo.GetForumThreadsSince(forum.Slug, since, limit, desc)
	}
	if err != nil {
		return nil, err
	}
	return threads, nil
}

func (u *Usecase) GetForumUsers(forum, since string, limit int, desc bool) (model.Users, error) {
	return u.repo.GetForumUsers(forum, since, limit, desc)
}

func (u *Usecase) VoteForThread(threadSlugOrID string, vote view.Vote) (*model.Thread, error) {
	thread, err := u.repo.GetThreadBySlugOrID(threadSlugOrID)
	if err != nil {
		return nil, err
	}
	userNick, err := u.repo.GetUserNickname(vote.Nickname)
	if err != nil {
		return nil, err
	}
	newVotes, err := u.repo.AddThreadVote(thread, userNick, vote.Voice)
	thread.Votes = newVotes
	return thread, err
}

func (u *Usecase) GetThread(threadSlugOrID string) (*model.Thread, error) {
	return u.repo.GetThreadBySlugOrID(threadSlugOrID)
}

func (u *Usecase) GetThreadPosts(threadSlugOrID string, limit int, since *int, sort string, desc bool) (model.Posts, error) {
	thread, err := u.repo.GetThreadFieldsBySlugOrID("id", threadSlugOrID)
	if err != nil {
		return nil, err
	}
	return u.repo.GetThreadPosts(thread.ID, limit, since, sort, desc)
}

type postDetails struct {
	Post   *model.Post
	Author *model.User
	Forum  *model.Forum
	Thread *model.Thread
}

func (u *Usecase) GetPostDetails(id int, related []string) (*postDetails, error) {
	post, err := u.repo.GetPostByID(id)
	if err != nil {
		return nil, err
	}
	details := postDetails{Post: post}
	for _, r := range related {
		switch r {
		case "user":
			details.Author, err = u.repo.GetUserByNickname(post.Author)
		case "forum":
			details.Forum, err = u.repo.GetForumBySlug(post.Forum)
		case "thread":
			details.Thread, err = u.repo.GetThreadByID(post.Thread)
		}
		if err != nil {
			return nil, err
		}
	}
	return &details, nil
}

func (u *Usecase) UpdatePost(id int, message string) (*model.Post, error) {
	return u.repo.UpdatePostMessage(id, message)
}

func (u *Usecase) GetStatus() (s view.Status, err error) {
	forum, err := u.repo.CountForums()
	if err != nil {
		return
	}
	post, err := u.repo.CountPosts()
	if err != nil {
		return
	}
	thread, err := u.repo.CountThreads()
	if err != nil {
		return
	}
	user, err := u.repo.CountUsers()
	if err != nil {
		return
	}
	s = view.Status{
		Forum:  forum,
		Post:   post,
		Thread: thread,
		User:   user,
	}
	return
}

func (u *Usecase) Clear() error {
	return u.repo.Clear()
}

