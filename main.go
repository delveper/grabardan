package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/delveper/env"
)

func main() {
	u := flag.String("u", "", "course URL")
	flag.Parse()

	if *u == "" {
		log.Println("Specify course URL.")
		os.Exit(1)
	}

	if err := env.LoadVars(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if err := FetchMediaFromCourse(*u); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func FetchMediaFromCourse(u string) error {
	const mediaDir = "media"

	course, err := parseCourse(u)
	if err != nil {
		return fmt.Errorf("parsing course: %w", err)
	}

	course.Title = strings.Replace(course.Title, ":", " -", -1)

	log.Printf("%+v", *course)

	if err := os.MkdirAll(path.Join(mediaDir, course.Title), 0666); err != nil {
		return fmt.Errorf("creating media dir: %w", err)
	}

	for _, chapter := range course.Chapters {
		log.Printf("Processing chapter: %v\n", chapter.Title)

		chapter.Title = strings.Replace(chapter.Title, ":", " -", -1)
		if err := os.MkdirAll(path.Join(mediaDir, course.Title, chapter.Title), 0666); err != nil {
			return fmt.Errorf("creating media subdir: %w", err)
		}

		for _, lesson := range chapter.Lessons {
			lesson.Title = strings.Replace(lesson.Title, ":", " -", -1)

			log.Printf("Processign lesson: %+v", lesson)

			if lesson.Slug == "" {
				log.Printf("Skipping lesson due empty slug: %+v", lesson)
				continue
			}

			asset, err := getBestAsset(path.Clean(path.Join("https://", hostURL, lesson.Slug)))
			if errors.Is(err, context.DeadlineExceeded) {
				log.Printf("Timeout: %s\n", u)
				continue
			}

			if err != nil {
				return fmt.Errorf("getting best asset: %w", err)
			}

			log.Printf("Best asset is: %+v\n", asset)

			media, err := downloadFile(asset.URL)
			if err != nil {
				return fmt.Errorf("processing lesson %+v: %w", lesson, err)
			}

			dst := path.Join(mediaDir, course.Title, chapter.Title, lesson.Title+".mp4")
			if err := saveFile(media, dst); err != nil {
				return fmt.Errorf("saving file from lesson %+v to path %v: %w", lesson, dst, err)
			}

			log.Println(" - Success")
		}
	}

	return nil
}
