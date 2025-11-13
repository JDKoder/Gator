package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/JDKoder/Gator/internal/config"
	"github.com/JDKoder/Gator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	//var conf config.Config = config.Config{DbURL: "foobar"}
	readConfig := config.Read()
	//initialize the db
	dbUrl := readConfig.DbURL
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("Unable to open database url", dbUrl)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	//set config and db into state
	State := state{config: &readConfig, db: dbQueries}
	commands := commands{commands: make(map[string]func(*state, command) error)}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerGetUsers)
	commands.register("agg", handlerAggregate)
	commands.register("feeds", handlerListFeeds)
	commands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commands.register("follow", middlewareLoggedIn(handlerFollow))
	commands.register("following", middlewareLoggedIn(handlerFollows))
	commands.register("unfollow", middlewareLoggedIn(handlerUnfollow))

	if len(os.Args) < 2 {
		log.Printf("Invalid argument length.  Expecting command argument.")
		os.Exit(1)
	}
	userCommand := command{name: strings.ToLower(os.Args[1]), args: os.Args[2:]}
	runError := commands.run(&State, userCommand)
	if runError != nil {
		os.Exit(1)
	}
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.config.CurrentUsername)
		if err != nil {
			log.Fatalf("no user found with name %s", cmd.args[0])
			os.Exit(1)
		}
		return handler(s, cmd, user)
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("the login handler expects a single argument, the username")
	}
	log.Printf("Logging in as user %s", cmd.args[0])
	user, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		log.Printf("no user found with name %s", cmd.args[0])
		os.Exit(1)
	}
	s.config.SetUser(user.Name)
	fmt.Printf("User has been set %s\n", s.config.CurrentUsername)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("the register handler expects a single argument, the username")
	}
	log.Printf("Registering user %s\n", cmd.args[0])

	newUser := database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.args[0]}
	user, err := s.db.CreateUser(context.Background(), newUser)
	if err != nil {
		log.Fatal("Duplicate user error.")
		os.Exit(1)
	}
	s.config.SetUser(cmd.args[0])
	log.Printf("User %s registered with id %v", cmd.args[0], user.ID)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		log.Fatalf("Failed to reset users table: %+v", err)
		os.Exit(1)
	}
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		log.Fatalf("Failed to get users: %+v", err)
		os.Exit(1)
	}
	var str string
	for _, user := range users {
		str = "%s\n"
		if user.Name == s.config.CurrentUsername {
			str = "%s (current)\n"
		}
		fmt.Printf(str, user.Name)
	}
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	var rssFeed RSSFeed
	if err != nil {
		return &rssFeed, err
	}
	request.Header.Set("User-Agent", "gator")
	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		log.Fatal("request to url", feedURL, "could not be completed")
		os.Exit(1)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal("failed to read response body.")
		os.Exit(1)
	}
	err = xml.Unmarshal(body, &rssFeed)
	if err != nil {
		log.Fatal("unable to marshal RSSFeed")
		os.Exit(1)
	}

	return &rssFeed, nil
}

func handlerAggregate(s *state, cmd command) error {
	var aggRefreshDur string = "1m"
	if len(cmd.args) < 1 {
		log.Printf("should provide a duration argument.  ie. 1m or 30s or 1h")
		log.Printf("Using default duration 1m")
	} else {
		aggRefreshDur = cmd.args[0]
	}
	dur, durationErr := time.ParseDuration(aggRefreshDur)
	if durationErr != nil {
		log.Fatalf("%w", durationErr)
	}
	ticker := time.NewTicker(dur)
	for ; ; <-ticker.C {
		rssFeed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
		if err != nil {
			log.Fatal("Couldn't fetch rss feed.")
			os.Exit(1)
		}
		rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
		rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
		for i, item := range rssFeed.Channel.Item {
			rssFeed.Channel.Item[i].Title = html.UnescapeString(item.Title)
			rssFeed.Channel.Item[i].Description = html.UnescapeString(item.Description)
		}
		fmt.Printf("%+v\n", rssFeed)
	}

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		log.Fatalf("Not enough arguments")
		os.Exit(1)
	}
	rssFeedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	feed, erro := s.db.CreateFeed(context.Background(), rssFeedParams)
	if erro != nil {
		log.Fatalf("could not create feed with values: %+v\n", rssFeedParams)
		os.Exit(1)
	}
	addFeedFollow(user.ID, feed.ID, s)
	fmt.Printf("%+v\n", feed)
	return nil
}

func handlerListFeeds(s *state, cmd command) error {
	feedsList, erro := s.db.GetUserFeeds(context.Background())
	if erro != nil {
		log.Fatalf("Failed to GetUserFeeds\n")
		os.Exit(1)
	}
	for _, feeds := range feedsList {
		fmt.Printf("%s\t%s\t%s\n", feeds.Name, feeds.Url, feeds.Username)
	}
	return nil
}

// takes a single url argument and creates a new feed follow record for the current user. It should print the name of the feed and the current user once the record is created (which the query we just made should support). You'll need a query to look up feeds by URL.
func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		log.Fatalln("Missing argument: URL")
		os.Exit(1)
	}
	feedAtUrl, feedErr := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if feedErr != nil {
		log.Fatalf("no feed found for url")
		os.Exit(1)
	}
	createdRow := addFeedFollow(user.ID, feedAtUrl.ID, s)
	fmt.Printf("%s\t%s\n", createdRow.UserName, createdRow.FeedName)
	return nil
}

func addFeedFollow(userId uuid.UUID, feedId uuid.UUID, s *state) database.CreateFeedFollowRow {
	feedFollow := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		UserID:    userId,
		FeedID:    feedId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	createdRow, createError := s.db.CreateFeedFollow(context.Background(), feedFollow)
	if createError != nil {
		log.Fatalf("failed to create feed follow record for userid %s and feedid %s", userId, feedId)
		os.Exit(1)
	}
	return createdRow
}

func handlerFollows(s *state, cmd command, user database.User) error {
	userFeeds, querror := s.db.GetFeedFollowsForUser(context.Background(), user.Name)
	if querror != nil {
		log.Fatalf("Unable to getfeedfollowsforuser with username %s", user.Name)
		os.Exit(1)
	}
	for _, userFeed := range userFeeds {
		fmt.Println(userFeed.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		log.Fatal("expecting 1 arg: feed url")
		os.Exit(1)
	}
	feed, feedErr := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if feedErr != nil {
		log.Fatalf("invalid feed url: %s", cmd.args[0])
		os.Exit(1)
	}
	deleteArgs := database.DeleteFeedFollowsByUserAndFeedParams{UserID: user.ID, FeedID: feed.ID}
	deleteErr := s.db.DeleteFeedFollowsByUserAndFeed(context.Background(), deleteArgs)
	if deleteErr != nil {
		log.Fatalf("unabled to delete feed with url %s", cmd.args[0])
	}
	return nil
}

func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Fatalf("Could not get next feed to fetch: %w", err)
		os.Exit(1)
	}
	if &feed == nil {
		log.Println("No feeds available to scrape.")
		return
	}
	markErr := s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:            feed.ID})
	if markErr != nil {
		log.Fatalf("Unable to mark feed id [%s] fetched", feed.ID)
		os.Exit(1)
	}

	rssFeed, fetchErr := fetchFeed(context.Background(), feed.Url)
	if fetchErr != nil {
		log.Fatal("Unable to fetch feed at url %s", feed.Url)
		os.Exit(1)
	}
	for _, item := range rssFeed.Channel.Item {
		fmt.Printf("%s\n", item.Title)
	}
}
