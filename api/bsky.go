package api

import (
	"context"
	"fmt"

	bsky "github.com/whyrusleeping/gosky/api/bsky"
	"github.com/whyrusleeping/gosky/xrpc"
)

type BskyApp struct {
	C *xrpc.Client
}

type PostEntity struct {
	Index *TextSlice `json:"index" cborgen:"index"`
	Type  string     `json:"type" cborgen:"type"`
	Value string     `json:"value" cborgen:"value"`
}

type TextSlice struct {
	Start int64 `json:"start" cborgen:"start"`
	End   int64 `json:"end" cborgen:"end"`
}

type ReplyRef struct {
	Root   PostRef `json:"root" cborgen:"root"`
	Parent PostRef `json:"parent" cborgen:"parent"`
}

type PostRecord struct {
	Type      string        `json:"$type,omitempty" cborgen:"$type"`
	Text      string        `json:"text" cborgen:"text"`
	Entities  []*PostEntity `json:"entities,omitempty" cborgen:"entities"`
	Reply     *ReplyRef     `json:"reply,omitempty" cborgen:"reply"`
	CreatedAt string        `json:"createdAt" cborgen:"createdAt"`
}

type PostRef struct {
	Uri string `json:"uri"`
	Cid string `json:"cid"`
}

type GetTimelineResp struct {
	Cursor string     `json:"cursor"`
	Feed   []FeedItem `json:"feed"`
}

type FeedItem struct {
	Uri        string      `json:"uri"`
	Cid        string      `json:"cid"`
	Author     *User       `json:"author"`
	RepostedBy *User       `json:"repostedBy"`
	MyState    *MyState    `json:"myState"`
	Record     interface{} `json:"record"`
}

type MyState struct {
	Repost   string `json:"repost"`
	Upvote   string `json:"upvote"`
	Downvote string `json:"downvote"`
}

type Declaration struct {
	Cid       string `json:"cid"`
	ActorType string `json:"actorType"`
}

type User struct {
	Did         string       `json:"did"`
	Handle      string       `json:"handle"`
	DisplayName string       `json:"displayName"`
	Declaration *Declaration `json:"declaration"`
}

func (b *BskyApp) FeedGetTimeline(ctx context.Context, algo string, limit int, before *string) (*bsky.FeedGetTimeline_Output, error) {
	params := map[string]interface{}{
		"algorithm": algo,
		"limit":     limit,
	}

	if before != nil {
		params["before"] = *before
	}

	var out bsky.FeedGetTimeline_Output
	if err := b.C.Do(ctx, xrpc.Query, encJson, "app.bsky.feed.getTimeline", params, nil, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (b *BskyApp) FeedGetAuthorFeed(ctx context.Context, author string, limit int, before *string) (*bsky.FeedGetAuthorFeed_Output, error) {
	params := map[string]interface{}{
		"author": author,
		"limit":  limit,
	}

	if before != nil {
		params["before"] = *before
	}

	var out bsky.FeedGetAuthorFeed_Output
	if err := b.C.Do(ctx, xrpc.Query, encJson, "app.bsky.feed.getAuthorFeed", params, nil, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

type GSADeclaration struct {
	Cid       string `json:"cid"`
	ActorType string `json:"actorType"`
}

type GetSuggestionsActor struct {
	Did         string          `json:"did"`
	Declaration *GSADeclaration `json:"declaration"`
	Handle      string          `json:"handle"`
	DisplayName string          `json:"displayName"`
	Description string          `json:"description"`
	IndexedAt   string          `json:"indexedAt"`
}

type GetSuggestionsResp struct {
	Cursor string                `json:"cursor"`
	Actors []GetSuggestionsActor `json:"actors"`
}

func (b *BskyApp) ActorGetSuggestions(ctx context.Context, limit int, cursor *string) (*GetSuggestionsResp, error) {
	params := map[string]interface{}{
		"limit": limit,
	}

	if cursor != nil {
		params["cursor"] = *cursor
	}

	var out GetSuggestionsResp
	if err := b.C.Do(ctx, xrpc.Query, "", "app.bsky.actor.getSuggestions", params, nil, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (b *BskyApp) FeedSetVote(ctx context.Context, subject *PostRef, direction string) error {
	body := map[string]interface{}{
		"subject":   subject,
		"direction": direction,
	}

	var out map[string]interface{}
	if err := b.C.Do(ctx, xrpc.Procedure, encJson, "app.bsky.feed.setVote", nil, body, &out); err != nil {
		return err
	}

	fmt.Println(out)

	return nil
}

type GetFollowsResp struct {
	Subject *User  `json:"subject"`
	Cursor  string `json:"cursor"`
	Follows []User `json:"follows"`
}

func (b *BskyApp) GraphGetFollows(ctx context.Context, user string, limit int, before *string) (*GetFollowsResp, error) {
	params := map[string]interface{}{
		"user":  user,
		"limit": limit,
	}

	if before != nil {
		params["before"] = *before
	}

	var out GetFollowsResp
	if err := b.C.Do(ctx, xrpc.Query, "", "app.bsky.graph.getFollows", params, nil, &out); err != nil {
		return nil, err
	}

	return &out, nil
}
