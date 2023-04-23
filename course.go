package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/chromedp/chromedp"
	"github.com/delveper/gostruct"
)

const host = "courses.ardanlabs.com"

type Course struct {
	Slug    string    `xpath:"-"`
	Title   string    `xpath:"//h1[contains(@class, 'course-progress__title')]"`
	Lessons []Lesson  `xpath:"//li[@data-qa='content-item'][..]"`
	Date    time.Time `xpath:"-"`
}

type Lesson struct {
	Slug  string `json:"slug" xpath:"//a[contains(@class, 'content-item__link')]/@href"`
	Title string `json:"title" xpath:"//div[contains(@class, 'content-item__title')]/text()"`
}

func parseCourse(u string) (*Course, error) {
	body, err := parseHTMLBody(u)
	if err != nil {
		return nil, fmt.Errorf("parsing hmtl: %w", err)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("empty body")
	}

	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parsing html tree: %w", err)
	}

	course, err := gostruct.MakeFromHTMLNode[Course](doc, "xpath")
	if err != nil {
		return nil, fmt.Errorf("parsing course struct: %w", err)
	}

	course.Slug = u
	course.Date = time.Now()

	course.Lessons, err = gostruct.MakeManyFromHTMLNode[Lesson](doc, "//li[@data-qa='content-item']", "xpath")
	if err != nil {
		return nil, fmt.Errorf("parsing lessons: %w", err)
	}

	return &course, nil
}

func parseHTMLBody(u string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute/2)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	token, err := parseToken()
	if err != nil {
		return "", err
	}

	var body string
	if err := chromedp.Run(ctx,
		setCookie(*token),
		chromedp.Navigate(u),
		chromedp.Sleep(time.Minute/4),
		chromedp.InnerHTML("//body", &body, chromedp.BySearch),
	); err != nil {
		return "", err
	}

	return body, nil
}
