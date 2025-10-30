package opengraph

import (
	"context"
	"net/url"

	"github.com/otiai10/opengraph"
)

type Preview struct {
	Title string
	Image string
	URL   string
}

func FetchPreview(ctx context.Context, link string) (*Preview, error) {
	og, err := opengraph.Fetch(link)
	if err != nil {
		return nil, err
	}

	img := ""
	if len(og.Image) > 0 {
		img = makeAbs(link, og.Image[0].URL)
	}
	return &Preview{
		Title: og.Title,
		Image: img,
		URL:   og.URL.String(),
	}, nil
}

func makeAbs(baseURL, maybeRelative string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return maybeRelative
	}
	ref, err := url.Parse(maybeRelative)
	if err != nil {
		return maybeRelative
	}
	return base.ResolveReference(ref).String()
}
