# Gator

CLI for multi user RSS feed aggregation ( aggre gator ) stores user and feeds from given urls in local database.  Feeds can be continuously aggregated in a background process.  Users can follow specific feeds and browse their feeds on demand.

# Dependencies

* [PostgresSql](https://www.boot.dev/lessons/74bea1f2-19cd-4ea9-966e-e2ca9dd1dfa9) database installed and running
* [Go](https://webinstall.dev/golang/) 
* [goose](https://github.com/pressly/goose)

# Setup Steps

* Install Dependencies and setup a running postgresql server locally
* Clone the repo, open a command line and run the following from the root of the project
```go install```
```go build```
* Run goose up to setup database
```
cd .../Gator/sql/schema
goose postgres "postgres://postgres:postgres@localhost:5432/gator" up
```
* create the .gatorconfig.json dot file in your user's home directory
```~/.gatorconfig.json```
* Start the aggregator (recommend giving a duration that won't be flagged against the rss hosts ie. Greater than 15 seconds)
```Gator agg "1m"```
* Add a user

# Example gatorconfig.json

* db_url: gator is built to run with a postgres sql database.  Change this part to connect to the database you setup databaseusername:password@host postgres and the password is postgres
* current_user_name: you can provide a default here, or leave it blank for your first run.
```
{
    "db_url":"postgres://postgres:postgres@localhost:5432/gator?sslmode=disable",
    "current_user_name":"default_user"
}
```

# Example usage:

* start the aggregator in a command prompt passing the duration between scrapes 
```Gator agg "1m"```
* register a user (automatically logs the user in)
```Gator register [username]"```
* login given existing username (must be registered)
```Gator login [username]```
* Add a feed *NOTE: this does not automatically follow the feed for your user*
```Gator addFeed [Name/Alias] [URL]```
* View a list of added feeds
```Gator feeds```
* As your current user, follow a feed given the url
```Gator follow [feed_url]```
* List feeds the current user follows
```Gator following```
* Unfollow a feed given the url
```Gator unfollow [feed_url]```
* browse the most recent posts scraped on your followed feeds.  If no number is given, default will show 2 posts.
```Gator browse [number]```
* Show list of commands (no context, or description)
```Gator help```


