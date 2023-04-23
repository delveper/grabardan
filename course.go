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

const hostURL = "courses.ardanlabs.com"

type Course struct {
	Slug     string    `json:"slug" xpath:"-"`
	Title    string    `json:"title" xpath:"//h1[contains(@class, 'course-progress__title')]"`
	Chapters []Chapter `json:"chapters" xpath:"//div[contains(@class, 'chapter-item__container')][..]"`
	Date     time.Time `json:"date" xpath:"-"`
}

type Chapter struct {
	Title   string `json:"title" xpath:"//div[contains(@class, 'chapter-item__container')]//h2[contains(@class, 'chapter-item__title')]"`
	Lessons []Item `json:"lessons" xpath:"//ul[contains(@class, 'chapter-item__contents')][..]//li[@data-qa='content-item'][..]"`
}

type Item struct {
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

	const (
		chapterSel       = "//div[contains(@class, 'chapter-item__container')]"
		chapterLessonSel = "//ul[contains(@class, 'chapter-item__contents')]"
		lessonSel        = "//li[@data-qa='content-item']"
	)

	course.Chapters, err = gostruct.MakeManyFromHTMLNode[Chapter](doc, chapterSel, "xpath")
	if err != nil {
		return nil, fmt.Errorf("parsing chapters: %w", err)
	}

	for i, node := range htmlquery.Find(doc, "//div[contains(@class, 'chapter-item__container')]//h2[contains(@class, 'chapter-item__title')]") {
		course.Chapters[i].Title = htmlquery.InnerText(node)
	}

	for i := range course.Chapters {
		course.Chapters[i].Lessons, err = gostruct.MakeManyFromHTMLNode[Item](doc, lessonSel, "xpath")
		if err != nil {
			return nil, fmt.Errorf("parsing lessons: %w", err)
		}

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
