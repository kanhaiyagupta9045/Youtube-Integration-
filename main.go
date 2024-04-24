package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"youtube/service"
)

func main() {
	var channelName string

	fmt.Println("Enter channel name: ")
	fmt.Scanln(&channelName)

	youtubeService, err := service.InitService()
	if err != nil {
		log.Fatalf("Error creating YouTube client: %v", err)
	}

	call := youtubeService.Search.List([]string{"id", "snippet"}).Q(channelName).Type("channel").MaxResults(1)
	response, err := call.Do()
	if err != nil {
		log.Fatalf("Error calling YouTube API: %v", err)
	}

	if len(response.Items) == 0 {
		log.Fatalf("No channels found with name: %s", channelName)
	}

	var wg sync.WaitGroup
	for _, item := range response.Items {
		if item.Id.Kind == "youtube#channel" {
			channelID := item.Id.ChannelId
			channelTitle := item.Snippet.Title

			wg.Add(1)
			go func(channelID, channelTitle string) {
				defer wg.Done()

				channelCall := youtubeService.Channels.List([]string{"statistics"}).Id(channelID)
				channelResponse, err := channelCall.Do()
				if err != nil {
					log.Printf("Error retrieving channel details: %v", err)
					return
				}

				if len(channelResponse.Items) == 0 {
					log.Printf("No channel details found for ID: %s", channelID)
					return
				}

				channel := channelResponse.Items[0]
				if channel.Statistics == nil {
					log.Printf("Channel statistics not available for ID: %s", channelID)
					return
				}

				videoCount := channel.Statistics.VideoCount
				subscriberCount := channel.Statistics.SubscriberCount
				viewCount := channel.Statistics.ViewCount

				fmt.Printf("Channel '%s' (ID: %s) has %d subscribers, %d videos, and %d views\n", channelTitle, channelID, subscriberCount, videoCount, viewCount)

				videoCall := youtubeService.Search.List([]string{"id"}).Q(channelName).Type("video").MaxResults(1)
				videoResponse, err := videoCall.Do()
				if err != nil {
					log.Printf("Error calling YouTube API for videos: %v", err)
					return
				}

				for _, item := range videoResponse.Items {
					if item.Id.Kind == "youtube#video" {
						videoID := item.Id.VideoId

						videoDetailsCall := youtubeService.Videos.List([]string{"statistics", "contentDetails"}).Id(videoID)
						videoDetailsResponse, err := videoDetailsCall.Do()
						if err != nil {
							log.Printf("Error retrieving video details: %v", err)
							return
						}

						if len(videoDetailsResponse.Items) == 0 {
							log.Printf("No video details found for ID: %s", videoID)
							return
						}

						video := videoDetailsResponse.Items[0]
						durationStr := video.ContentDetails.Duration

						var duration string
						if durationStr != "" {
							durationParts := strings.Split(durationStr, "PT")
							if len(durationParts) > 1 {
								duration = durationParts[1]
							} else {
								duration = durationParts[0]
							}
						} else {
							duration = "Unknown"
						}

						fmt.Printf("Video ID: %s\n", videoID)
						fmt.Printf("Duration: %s\n", duration)

						if video.Statistics != nil {
							stats := video.Statistics
							fmt.Printf("Views: %d\n", stats.ViewCount)
							fmt.Printf("Likes: %d\n", stats.LikeCount)
							fmt.Printf("Dislikes: %d\n", stats.DislikeCount)
						} else {
							fmt.Println("Video statistics not available")
						}

						fmt.Println("Comments:")
						commentThreadsCall := youtubeService.CommentThreads.List([]string{"snippet", "replies"}).VideoId(videoID).MaxResults(100)
						commentThreadsResponse, err := commentThreadsCall.Do()
						if err != nil {
							log.Printf("Error retrieving video comments: %v", err)
							return
						}

						for _, commentThread := range commentThreadsResponse.Items {
							comment := commentThread.Snippet.TopLevelComment
							fmt.Printf("Author: %s\n", comment.Snippet.AuthorDisplayName)
							fmt.Printf("Comment: %s\n", comment.Snippet.TextDisplay)
							fmt.Printf("Likes: %d\n", comment.Snippet.LikeCount)

							if commentThread.Replies != nil && len(commentThread.Replies.Comments) > 0 {
								fmt.Println("Replies:")
								for _, reply := range commentThread.Replies.Comments {
									fmt.Printf("Author: %s\n", reply.Snippet.AuthorDisplayName)
									fmt.Printf("Comment: %s\n", reply.Snippet.TextDisplay)
									fmt.Printf("Likes: %d\n", reply.Snippet.LikeCount)
								}
							} else {
								fmt.Println("No replies found for this comment")
							}
							fmt.Println()
						}
						fmt.Println("----------------------------")
					}
				}
			}(channelID, channelTitle)
		}
	}

	wg.Wait()
}
