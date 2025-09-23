package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zrtgzrtg/gatorcli/internal/config"
	"github.com/zrtgzrtg/gatorcli/internal/database"
	"github.com/zrtgzrtg/gatorcli/rss"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}
type command struct {
	name string
	args []string
}
type commands struct {
	cMap map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("no arguments found in login command!\n")
	}
	name := cmd.args[0]
	//check if user exists
	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return err
	}

	err = s.cfg.SetUser(name)
	if err != nil {
		return err
	}
	fmt.Printf("User: %v has been set to config", name)
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	val, ok := c.cMap[cmd.name]
	if !ok {
		return fmt.Errorf("command with name: %v not found!", cmd.name)
	}
	err := val(s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cMap[name] = f
}
func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("no name found in register command")
	}
	timeNow := time.Now()
	sqlTimeNowStructObj := sql.NullTime{timeNow, true}
	usrName := cmd.args[0]
	usrParams := database.CreateUserParams{uuid.New(), sqlTimeNowStructObj, sqlTimeNowStructObj, usrName}

	// check if err is nil --> if yes it already exists
	_, err := s.db.GetUser(context.Background(), usrName)
	if err == nil {
		fmt.Printf("user with name: %v already exists in db!\n", usrName)
		os.Exit(1)
	}
	usr, err := s.db.CreateUser(context.Background(), usrParams)
	if err != nil {
		return err
	}
	s.cfg.SetUser(usrName)
	fmt.Printf("User was created with data:\n ID: %v \n created_at: %v \n updated_at: %v\n name: %v\n", usr.ID, usr.CreatedAt, usr.UpdatedAt, usr.Name)
	return nil
}
func handlerReset(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return err
	} else {
		os.Exit(0)
	}
	return nil
}
func handlerUsers(s *state, cmd command) error {
	//get list of users out of db
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	printList := []string{}
	for _, usr := range users {
		if usr.Name == s.cfg.Current_user_name {
			printList = append(printList, fmt.Sprintf("  * %v (current)\n", usr.Name))
		} else {
			printList = append(printList, fmt.Sprintf("  * %v\n", usr.Name))
		}
	}
	for _, p := range printList {
		fmt.Print(p)
	}
	return nil
}
func handlerAgg(s *state, cmd command) error {
	timeBetweenReqs := cmd.args[0]
	tim, err := time.ParseDuration(timeBetweenReqs)
	if err != nil {
		return err
	}
	fmt.Printf("Collecting feeds every %v\n", tim)
	tick := time.NewTicker(tim)
	for ; ; <-tick.C {
		feed, err := s.db.GetNextFeedToFetch(context.Background())
		if err != nil {
			return err
		}
		furl := feed.Url
		rssFeed, err := rss.FetchFeed(context.Background(), furl)
		if err != nil {
			return err
		}
		markParams := database.MarkFeedFetchedByIdParams{feed.ID, sql.NullTime{time.Now(), true}}
		_, err = s.db.MarkFeedFetchedById(context.Background(), markParams)
		if err != nil {
			return err
		}
		tim := sql.NullTime{time.Now(), true}
		postParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   tim,
			UpdatedAt:   tim,
			Title:       rssFeed.Channel.Title,
			Url:         furl,
			Description: rssFeed.Channel.Description,
			PublishedAt: time.Now(),
			FeedID:      feed.ID,
		}
		_, err = s.db.CreatePost(context.Background(), postParams)
		if !strings.Contains(err.Error(), "unique constraint") {
			return err
		}
	}

	return nil
}
func handlerAddFeed(s *state, cmd command) error {
	user, err := s.db.GetUser(context.Background(), s.cfg.Current_user_name)
	if err != nil {
		fmt.Println("a")
		return err
	}
	name := cmd.args[0]
	url := cmd.args[1]

	time := time.Now()
	feedID := uuid.New()
	feed := database.CreateFeedParams{
		ID:        feedID,
		CreatedAt: sql.NullTime{time, true},
		UpdatedAt: sql.NullTime{time, true},
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}
	feedRet, err := s.db.CreateFeed(context.Background(), feed)
	if err != nil {
		return err
	}
	feedFollowsParams := database.CreateFeedFollowsParams{
		ID:     uuid.New(),
		UserID: user.ID,
		FeedID: feedID,
	}
	_, err = s.db.CreateFeedFollows(context.Background(), feedFollowsParams)
	if err != nil {
		return err
	}
	fmt.Println(feedRet.ID, feedRet.CreatedAt, feedRet.UpdatedAt, feedRet.Name, feedRet.Url, feedRet.UserID)
	return nil
}
func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		usr, err := s.db.GetUserById(context.Background(), feed.UserID)
		if err != nil {
			return err
		}
		fmt.Println(feed.ID, feed.CreatedAt, feed.UpdatedAt, feed.Name, feed.Url, feed.UserID, usr.Name)
	}
	return nil
}
func handlerFollow(s *state, cmd command) error {
	usr, err := s.db.GetUser(context.Background(), s.cfg.Current_user_name)
	if err != nil {
		return err
	}
	usrID := usr.ID
	url := cmd.args[0]
	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return err
	}
	ret, err := s.db.CreateFeedFollows(context.Background(), database.CreateFeedFollowsParams{uuid.New(), usrID, feed.ID})
	if err != nil {
		return err
	}
	fmt.Println(ret.Username, ret.Feedname)
	return nil
}
func handlerFollowing(s *state, cmd command) error {
	usr, err := s.db.GetUser(context.Background(), s.cfg.Current_user_name)
	if err != nil {
		return err
	}
	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), usr.ID)
	if err != nil {
		return err
	}
	for _, feedFollow := range feedFollows {
		fmt.Println(feedFollow.Feedname)
	}

	return nil
}
func handlerUnfollow(s *state, cmd command) error {
	usr, err := s.db.GetUser(context.Background(), s.cfg.Current_user_name)
	if err != nil {
		return err
	}
	feedUrl := cmd.args[0]
	feed, err := s.db.GetFeedByUrl(context.Background(), feedUrl)
	if err != nil {
		return err
	}
	delParams := database.DeleteFeedFollowParams{usr.ID, feed.ID}
	_, err = s.db.DeleteFeedFollow(context.Background(), delParams)
	if err != nil {
		return err
	}
	return nil
}
func handlerScrapeFeeds(s *state, cmd command) error {
	nxtFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	markParams := database.MarkFeedFetchedByIdParams{nxtFeed.ID, sql.NullTime{time.Now(), true}}

	_, err = s.db.MarkFeedFetchedById(context.Background(), markParams)
	if err != nil {
		return err
	}
	rssFeed, err := rss.FetchFeed(context.Background(), nxtFeed.Url)
	for _, item := range rssFeed.Channel.Item {
		fmt.Println(item.Title)
	}

	return nil
}
func handlerBrowse(s *state, cmd command) error {
	user, err := s.db.GetUser(context.Background(), s.cfg.Current_user_name)
	if err != nil {
		return err
	}
	lim := 2
	if len(cmd.args) != 0 {
		lim, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return err
		}
	}
	getParams := database.GetPostsForUserParams{user.ID, int32(lim)}
	posts, err := s.db.GetPostsForUser(context.Background(), getParams)
	if err != nil {
		return err
	}
	for _, post := range posts {
		fmt.Println(post.Title)
	}

	return nil
}
