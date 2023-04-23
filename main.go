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
	const defaultCourse = "https://courses.ardanlabs.com/courses/take/ultimate-go-web-services-4-0"
	u := flag.String("u", defaultCourse, "course URL")
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

	for _, lesson := range course.Lessons {
		lesson.Title = strings.Replace(lesson.Title, ":", " -", -1)

		log.Printf("Processign lesson: %+v", lesson)

		if lesson.Slug == "" {
			log.Printf("Skipping lesson due empty slug: %+v", lesson)
			continue
		}

		p := path.Clean(path.Join("https://", host, lesson.Slug))
		asset, err := getBestAsset(p)
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

		dst := path.Join(mediaDir, course.Title, lesson.Title+".mp4")
		if err := saveFile(media, dst); err != nil {
			return fmt.Errorf("saving lesson %+v to path %v: %w", lesson, dst, err)
		}

		log.Println(" - Success")
	}

	return nil
}
