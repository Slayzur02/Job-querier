package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
	"github.com/chromedp/chromedp"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load(".env")
}

func main() {
	// initiate context for bot
	headless := false

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", headless),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	all_comps := map[string]int{}

	// scraping from levels
	// levelSrc := getLevelsHTML(ctx)
	// levelCompanies := levelScrape(levelSrc)
	// for _, comp := range levelCompanies {
	// 	all_comps[comp] = 1
	// 	continue
	// }

	// scraping from pittsc website
	// resp, err := soup.Get("https://github.com/pittcsc/Summer2022-Internships")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// pittCompanies := pittScrape(resp)

	// for _, comp := range pittCompanies {
	// 	if _, ok := all_comps[comp]; ok {
	// 		continue
	// 	} else {
	// 		all_comps[comp] = 1
	// 	}
	// }

	// time.Sleep(1 * time.Second)

	// scrape from discord
	discordSrc := getDiscordHTML(ctx)
	discordScrape(discordSrc)

	// turn map into arr

	allCompSlice := []string{}
	for comp := range all_comps {
		allCompSlice = append(allCompSlice, comp)
	}

	// for _, c := range allCompSlice {
	// 	fmt.Println(c)
	// }
}

func getDiscordHTML(ctx context.Context) string {
	var discordSrc string
	discord_username := os.Getenv("discord_username")
	discord_password := os.Getenv("discord_password")
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://discord.com/channels/@me"),
		// chromedp.OuterHTML(`body`, &discordLogin),

		chromedp.SendKeys(`//input[@name="email"]`, discord_username, chromedp.BySearch),
		chromedp.SendKeys(`//input[@name="password"]`, discord_password, chromedp.BySearch),
		chromedp.Click(`//button[@type="submit"]`, chromedp.BySearch),
	)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(3 * time.Second)

	err = chromedp.Run(ctx,
		chromedp.Click(`//div[@role="treeitem" and contains(@aria-label, "cscareers.dev")]`, chromedp.BySearch),
		chromedp.Navigate(`https://discord.com/channels/698366411864670250/856027992701927444`),
		// chromedp.Navigate("https://discord.com/channels/698366411864670250/856027992701927444"),
	)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(4 * time.Second)

	amountScrollUp := 10

	for i := 0; i < amountScrollUp; i++ {
		err = chromedp.Run(ctx,
			chromedp.ScrollIntoView(`//ol/li[1]`, chromedp.BySearch),
		)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(1 * time.Second)
	}

	err = chromedp.Run(ctx,
		chromedp.OuterHTML(`//ol`, &discordSrc, chromedp.BySearch),
	)

	if err != nil {
		log.Fatal(err)
	}

	return discordSrc
}

func discordScrape(source string) {
	doc := soup.HTMLParse(source)
	listItems := doc.FindAll("li")

	fmt.Println(len(listItems))
	for _, li := range listItems {
		listAttributes := li.Attrs()
		if _, ok := listAttributes["id"]; !ok {
			continue
		}
		// has id for list elements
		text := li.Children()[0].Children()[0].Find("div").Text()
		if strings.Contains(text, "Please follow") ||
			strings.Contains(text, "Congrats!!!") ||
			strings.Contains(text, "Company is not recognized") {
			continue
		}
		fmt.Println(text)
	}
}

func getLevelsHTML(ctx context.Context) string {
	var levelSrc string

	err := chromedp.Run(ctx,
		chromedp.Navigate(`https://www.levels.fyi/internships/?track=Software%20Engineer&timeframe=2022%20%2F%202021`),
		// chromedp.ScrollIntoView(`table`),
		// chromedp.WaitVisible(`table`),
		chromedp.OuterHTML(`body`, &levelSrc),
	)
	if err != nil {
		log.Fatal(err)
	}

	return levelSrc
}

func levelScrape(source string) (openSpots []string) {
	doc := soup.HTMLParse(source)
	rows := doc.Find(`tbody`).FindAll("tr")

	for _, r := range rows {
		if len(r.Children()) < 4 {
			continue
		}
		comp := r.Find("h6").Text()
		status := r.Children()[3].Find("a").Text()
		statusLower := strings.ToLower(status)

		if strings.Contains(statusLower, "apply") {
			openSpots = append(openSpots, strings.ToLower(strings.Trim(comp, " \n")))
		}
	}

	return openSpots
}

func pittScrape(source string) (openSpots []string) {
	doc := soup.HTMLParse(source)
	rows := doc.Find("tbody").FindAll("tr")

	for _, r := range rows {
		// get the first tag - del or a
		res := r.Find("td").Children()[0]

		attributes := res.Attrs()
		if _, ok := attributes["href"]; ok {
			openSpots = append(openSpots, strings.ToLower(strings.Trim(res.Text(), " ][\n")))
		}
	}

	return openSpots
}
