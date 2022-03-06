package flex

import (
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/roaris/news-ticker/newsapi"
)

func NewBubbleContainer(article newsapi.Article) linebot.BubbleContainer {
	return linebot.BubbleContainer{
		Type: "bubble",
		Size: "kilo",
		Header: &linebot.BoxComponent{
			Type:   "box",
			Layout: "vertical",
			Contents: []linebot.FlexComponent{
				&linebot.TextComponent{
					Type: "text",
					Text: article.Title,
					Wrap: true,
				},
			},
		},
		Hero: &linebot.ImageComponent{
			Type: "image",
			URL:  article.UrlToImage,
			Size: "5xl",
		},
		Footer: &linebot.BoxComponent{
			Type:   "box",
			Layout: "vertical",
			Contents: []linebot.FlexComponent{
				&linebot.ButtonComponent{
					Type: "button",
					Action: &linebot.URIAction{
						Label: "記事を読む",
						URI:   article.Url,
					},
				},
			},
		},
	}
}

func NewCaroucelContainer(contents []*linebot.BubbleContainer) linebot.CarouselContainer {
	return linebot.CarouselContainer{
		Type:     "caroucel",
		Contents: contents,
	}
}
