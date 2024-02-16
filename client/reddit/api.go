package reddit

import "fmt"

type RedditClientInterface interface {
	Submission(submissionId string)
	Subreddit(subreddit string)
	Submit(title string, text string)
}

type RedditClient struct{}

func (redditClient *RedditClient) Submission(submissionId string) {
	fmt.Printf("Submission. submissionId: %s\n", submissionId)
}

func (redditClient *RedditClient) Subreddit(subreddit string) {
	fmt.Printf("Subreddit. subreddit: %s\n", subreddit)
}

func (redditClient *RedditClient) Submit(title string, text string) {
	fmt.Printf("Submit. title: %s, text: %s\n", title, text)
}
