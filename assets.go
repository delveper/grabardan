package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type Token struct {
	Value   string    `env:"TOKEN" key:"remember_user_token"`
	Expires time.Time `env:"TOKEN_EXPIRATION" key:"-"`
}

func getBestAsset(u string) (*Asset, error) {
	u, err := getAssetSource(u)
	if err != nil {
		return nil, fmt.Errorf("getting asset source URL: %w", err)
	}

	b, err := downloadFile(u)
	if err != nil {
		return nil, fmt.Errorf("downloading file: %w", err)
	}

	j := parseJSONP(b)

	var jsond JSOND
	if err := json.Unmarshal(j, &jsond); err != nil {
		return nil, fmt.Errorf("unmarshaling JSOND: %s: %w", string(j), err)
	}

	var best Asset
	for _, asset := range jsond.Media.Assets {
		if asset.Size > best.Size {
			best = asset
		}
	}

	return &best, nil
}

func getAssetSource(u string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute/2)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		return "", err
	}

	ch := make(chan *network.EventResponseReceived, 100)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if ev, ok := ev.(*network.EventResponseReceived); ok {
			ch <- ev
		}
	})

	token, err := parseToken()
	if err != nil {
		return "", err
	}

	if err := chromedp.Run(ctx,
		setCookie(*token),
		chromedp.Navigate(u),
	); err != nil {
		return "", err
	}

	for {
		select {
		case ev := <-ch:
			if ok, err := path.Match("https://fast.wistia.com/embed/medias/*.jsonp", ev.Response.URL); ok && err == nil {
				return ev.Response.URL, nil
			}

		case <-time.After(time.Minute / 4):
			return "", context.DeadlineExceeded
		}
	}
}

func setCookie(token Token) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		exp := cdp.TimeSinceEpoch(token.Expires)
		if err := network.SetCookie("remember_user_token", token.Value).
			WithExpires(&exp).
			WithDomain(hostURL).
			WithPath("/").
			WithHTTPOnly(true).
			WithURL("").
			WithSecure(true).
			Do(ctx); err != nil {
			return fmt.Errorf("could not set cookie")
		}
		return nil
	})
}

func parseJSONP(b []byte) []byte {
	re := regexp.MustCompile(`{"media":[^\n;]*`)
	re.Match(b)

	return re.Find(b)
}

func parseToken() (*Token, error) {
	val := os.Getenv("TOKEN")
	if val == "" {
		return nil, fmt.Errorf("missing token")
	}

	exp, _ := time.Parse(time.RFC3339, os.Getenv("TOKEN_EXPIRATION"))
	token := Token{
		Value:   val,
		Expires: exp,
	}

	return &token, nil
}
