package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

var config struct {
	Username       *string `json:"Username"`
	ConsumerKey    *string `json:"ConsumerKey"`
	ConsumerSecret *string `json:"ConsumerSecret"`
	AccessToken    *string `json:"AccessToken"`
	AccessSecret   *string `json:"AccessSecret"`
	DeleteTweets   *bool   `json:"DeleteTweets"`
	DeleteRetweets *bool   `json:"DeleteRetweets"`
	DeleteLikes    *bool   `json:"DeleteLikes"`
}

func main() {
	// Open our config file
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal("Cannot open config.json: ", err)
	}

	// Decode our config file into our struct
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Cannot load JSON config into config.struct: ", err)
	}

	configTest := reflect.ValueOf(config)
	configType := configTest.Type()
	for i := 0; i < configTest.NumField(); i++ {
		if configTest.Field(i).IsNil() {
			log.Fatal("Required JSON value was not defined: ", configType.Field(i).Name)
		}
	}

	fmt.Println("  Starting..")

	oauthConfig := oauth1.NewConfig(*config.ConsumerKey, *config.ConsumerSecret)
	token := oauth1.NewToken(*config.AccessToken, *config.AccessSecret)
	// http.Client will automatically authorize Requests
	httpClient := oauthConfig.Client(oauth1.NoContext, token)

	// twitter client
	client := twitter.NewClient(httpClient)

	// Delete tweets and retweets
	if *config.DeleteTweets || *config.DeleteRetweets {
		fmt.Println("    Deleting tweets and retweets..")
		for true {
			tweets, _, err := client.Timelines.UserTimeline(&twitter.UserTimelineParams{
				ScreenName:      *config.Username,
				Count:           20,
				TrimUser:        twitter.Bool(true),
				IncludeRetweets: twitter.Bool(true),
			})

			if err != nil {
				log.Fatal("Failed to get tweets: ", err)
			}

			if len(tweets) == 0 {
				fmt.Println("    No (re)tweets remain.")
				break
			}

			for _, tweet := range tweets {
				if tweet.Retweeted {
					if *config.DeleteRetweets {
						fmt.Printf("      Deleting: retweeted id: %v\n", tweet.ID)
						params := &twitter.StatusUnretweetParams{TrimUser: twitter.Bool(true)}
						client.Statuses.Unretweet(tweet.ID, params)
					}
				} else {
					if *config.DeleteTweets {
						fmt.Printf("      Deleting: tweeted id: %v\n", tweet.ID)
						params := &twitter.StatusDestroyParams{TrimUser: twitter.Bool(true)}
						client.Statuses.Destroy(tweet.ID, params)
					}
				}
				time.Sleep(10 * time.Second)
			}
		}
	}

	// Delete likes
	if *config.DeleteLikes {
		fmt.Println("    Deleting likes..")
		for true {
			tweets, _, err := client.Favorites.List(&twitter.FavoriteListParams{
				ScreenName: *config.Username,
				Count:      20,
			})

			if err != nil {
				log.Fatal("Failed to get tweets: ", err)
			}

			if len(tweets) == 0 {
				fmt.Println("    No likes remain.")
				break
			}

			for _, tweet := range tweets {
				fmt.Printf("      Deleting: liked id: %v\n", tweet.ID)
				params := &twitter.FavoriteDestroyParams{ID: tweet.ID}
				client.Favorites.Destroy(params)
				time.Sleep(10 * time.Second)
			}
		}
	}

	// Done
	fmt.Println("  Done.")
}
