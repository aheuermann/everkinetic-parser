package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Exercise struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Steps           []string `json:"steps"`
	Type            string   `json:"type"`
	PrimaryMuscle   []string `json:"primary"`
	SecondaryMuscle []string `json:"secondary"`
	Equipment       []string `json:"equipment"`
}

var (
	count  = 0
	outDir = ""
)

func ScrapeExercise(page string) (Exercise, error) {
	count++
	ret := Exercise{}
	res, err := http.Get(page)
	if err != nil {
		panic(err)
	}
	if res.StatusCode != 200 {
		panic(fmt.Sprintf("Page:%s ResponseCode:%d", page, res.StatusCode))
	}

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		panic(err)
	}

	ret.Name = strings.TrimSpace(doc.Find(".entry-title a").Text())
	content := doc.Find(".exercise-entry-content")
	ret.Description = content.Find("p").First().Text()
	ret.Steps = []string{}
	content.Find("ol li").Each(func(i int, s *goquery.Selection) {
		ret.Steps = append(ret.Steps, strings.TrimSpace(s.Text()))
	})
	taxomolies := content.Find(".exercise-taxonomies li")
	typ := taxomolies.First()
	primary := typ.Next()
	secondary := primary.Next()
	equipment := secondary.Next()
	ret.Type = strings.TrimSpace(typ.Find("a").Text())
	ret.PrimaryMuscle = []string{}
	primary.Find("a").Each(func(i int, s *goquery.Selection) {
		ret.PrimaryMuscle = append(ret.PrimaryMuscle, strings.TrimSpace(s.Text()))
	})
	ret.SecondaryMuscle = []string{}
	secondary.Find("a").Each(func(i int, s *goquery.Selection) {
		ret.SecondaryMuscle = append(ret.SecondaryMuscle, strings.TrimSpace(s.Text()))
	})
	ret.Equipment = []string{}
	equipment.Find("a").Each(func(i int, s *goquery.Selection) {
		ret.Equipment = append(ret.Equipment, strings.TrimSpace(s.Text()))
	})

	content.Find(".download-exercise-images li").First().Next().Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			panic("...")
		}
		res, err := http.Get(href)
		if err != nil {
			panic(err)
		}
		if res.StatusCode != 200 {
			fmt.Printf("Error:%s - %s\n", page, href)
		}

		defer res.Body.Close()
		out, err := os.Create(filepath.Join(outDir, fmt.Sprintf("%d_%d.png", count, i)))
		if err != nil {
			panic(err)
		}
		defer out.Close()
		io.Copy(out, res.Body)
	})
	return ret, nil
}

func Crawl() {
	pages := []string{}
	for n := 1; true; n++ {
		page := fmt.Sprintf("http://db.everkinetic.com/page/%d/", n)
		res, err := http.Get(page)
		if err != nil {
			panic(err)
		}
		if res.StatusCode == 404 && n > 1 {
			break
		}
		if res.StatusCode != 200 {
			panic(fmt.Sprintf("Page:%s ResponseCode:%d", page, res.StatusCode))
		}

		doc, err := goquery.NewDocumentFromResponse(res)
		if err != nil {
			panic(err)
		}

		doc.Find(".exercise").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Find("a").First().Attr("href")
			if !exists {
				panic(fmt.Sprintf("link not found. Page:%s #%d", n, i))
			}

			pages = append(pages, href)
		})
	}
	exercises := make([]Exercise, len(pages))
	// var wg sync.WaitGroup
	// wg.Add(len(pages))
	for i, page := range pages {
		fmt.Printf("%d/%d - Scraping:%s\n", i+1, len(pages), page)
		//go func(idx int) {
		exercise, err := ScrapeExercise(page)
		if err != nil {
			panic(err)
		}
		exercises[i] = exercise
		fmt.Printf("%d/%d - Done\n", i, len(pages))
		//wg.Done()
		//}(i)
	}
	// wg.Wait()
	fmt.Printf("%#v", exercises)
	b, err := json.Marshal(exercises)
	if err != nil {
		panic(err)
	}
	os.Create(filepath.Join(outDir, "data.json"))

	if err = ioutil.WriteFile(filepath.Join(outDir, "data.json"), b, os.ModePerm); err != nil {
		panic(err)
	}
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	outDir = filepath.Join(wd, "out")
	os.Mkdir(outDir, os.ModePerm) //ignore error

	Crawl()
}
