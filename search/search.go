package search

import (
	"code.google.com/p/google-api-go-client/youtube/v3"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"sync"
)

var (
	maxResults int64 = 5  //One should be okay
)

func TopResult(keyword string) string {


	service := getService()

	// Make the API call to YouTube.
	call := service.Search.List("id,snippet").
		Q(keyword).
		MaxResults(maxResults).
		Type("video") //Not sure
	response, err := call.Do()
	if err != nil {
		log.Fatalf("Error making search API call: %v", err)
	}

	// Group video, channel, and playlist results in separate lists.
	//videos := make(map[string]string)
	//channels := make(map[string]string)
	//playlists := make(map[string]string)

	// Iterate through each item and add it to the correct list.
	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			return item.Id.VideoId
		//case "youtube#channel":
		//	channels[item.Id.ChannelId] = item.Snippet.Title
		//case "youtube#playlist":
		//	playlists[item.Id.PlaylistId] = item.Snippet.Title
		}
	}
	return ""
	//fmt.Println("Videos", videos)
	//fmt.Println("Channels", channels)
	//fmt.Println("Playlists", playlists)
}

func getService() (service *youtube.Service) {
	//@TODO - ADD SYNC.ONCE
	var once sync.Once
	once.Do(func() {
		client, err := buildOAuthHTTPClient(youtube.YoutubeScope)
		if err != nil {
			log.Fatalf("Error building OAuth client: %v", err)
		}
		service, err = youtube.New(client)
		if err != nil {
			log.Fatalf("Error creating YouTube client: %v", err)
		}
	})
	return service
}
