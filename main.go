package main

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"
)

// Example of using chromedp to navigate to vesselfinder.
// Not much useful GPS information in the client though.
func main() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36"),
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var res string
	var okay bool
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.vesselfinder.com/vessels/details/1012610"),
		chromedp.AttributeValue("#djson", "data-json", &res, &okay, chromedp.NodeVisible),
	)
	if okay == false {
		log.Fatal("Unable to read attribute data")
	}
	if err != nil {
		log.Fatal(err)
	}

	log.Println(res)
}
