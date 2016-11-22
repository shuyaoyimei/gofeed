package sitemap

import (
	"fmt"
	"io"
	"strings"

	"github.com/mmcdole/goxpp"
	"github.com/shuyaoyimei/gofeed/extensions"
	"github.com/shuyaoyimei/gofeed/internal/shared"
)

// Parser is a Sitemap Parser
type Parser struct{}

// Parse parses an xml feed into an sitemap.Feed
func (sp *Parser) Parse(feed io.Reader) (*Feed, error) {
	p := xpp.NewXMLPullParser(feed, false, shared.NewReaderLabel)

	_, err := shared.FindRoot(p)
	if err != nil {
		return nil, err
	}
	return sp.parseRoot(p)
}

func (sp *Parser) parseRoot(p *xpp.XMLPullParser) (*Feed, error) {
	sitemapErr := p.Expect(xpp.StartTag, "urlset")
	if sitemapErr != nil {
		return nil, fmt.Errorf("%s", sitemapErr.Error())
	}
	// Items found in feed root
	// var channel *Feed
	channel := &Feed{}
	items := []*Item{}

	ver := sp.parseVersion(p)

	for {
		tok, err := shared.NextTag(p)
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {

			// Skip any extensions found in the feed root.
			if shared.IsExtension(p) {
				p.Skip()
				continue
			}

			name := strings.ToLower(p.Name)

			if name == "url" {
				item, feed, err := sp.parseItem(p)
				if err != nil {
					return nil, err
				}
				items = append(items, item)
				if channel == nil {
					if feed.Title != "" {
						channel.Title = feed.Title
					} else {
						channel.Title = "unkonow"
					}
					if feed.Language != "" {
						channel.Language = feed.Language
					} else {
						channel.Language = "unknow"
					}
				}
			} else {
				p.Skip()
			}
		}
	}

	sitemapErr = p.Expect(xpp.EndTag, "urlset")
	if sitemapErr != nil {
		return nil, fmt.Errorf("%s", sitemapErr.Error())
	}

	if channel == nil {
		channel = &Feed{}
		channel.Items = []*Item{}
	}

	if len(items) > 0 {
		channel.Items = append(channel.Items, items...)
	}

	channel.Version = ver
	return channel, nil
}

func (sp *Parser) parseVersion(p *xpp.XMLPullParser) (ver string) {
	name := strings.ToLower(p.Name)
	if name == "urlset" {
		ns := p.Attribute("xmlns")
		if ns == "http://www.sitemaps.org/schemas/sitemap/0.9" {
			ver = "0.9"
		} else {
			ver = "unknow"
		}
	} else {
		ver = "unknow"
	}
	return
}

func (sp *Parser) parseItem(p *xpp.XMLPullParser) (item *Item, feed *Feed, err error) {

	if err = p.Expect(xpp.StartTag, "url"); err != nil {
		return nil, nil, err
	}

	item = &Item{}
	feed = &Feed{}
	extensions := ext.Extensions{}

	for {
		tok, err := shared.NextTag(p)
		if err != nil {
			return nil, nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {

			name := strings.ToLower(p.Name)

			if shared.IsExtension(p) {
				ext, err := shared.ParseExtension(extensions, p)
				if err != nil {
					return nil, nil, err
				}
				item.Extensions = ext
			} else if name == "news" {
				result, err := sp.parseNews(p)
				//must change last code
				if err != nil {
					return nil, nil, err
				}
				item.Title = result.Title
				item.PubDate = result.PublicationDate
				date, err := shared.ParseDate(result.PublicationDate)
				if err == nil {
					utcDate := date.UTC()
					item.PubDateParsed = &utcDate
				}
				feed.Title = result.Name
				feed.Language = result.Language
			} else if name == "loc" {
				if len(item.Link) > 0 {
					continue
				}
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, nil, err
				}
				item.Link = result
			} else if name == "image" {
				result, err := sp.parseImage(p)
				if err != nil {
					return nil, nil, err
				}
				item.Image = result
			} else {
				// Skip any elements not part of the item spec
				p.Skip()
			}
		}
	}

	if len(extensions) > 0 {
		item.Extensions = extensions
	}

	if err = p.Expect(xpp.EndTag, "url"); err != nil {
		return nil, nil, err
	}

	return item, feed, nil
}

func (sp *Parser) parseNews(p *xpp.XMLPullParser) (news *News, err error) {
	if err = p.Expect(xpp.StartTag, "news"); err != nil {
		return nil, err
	}

	news = &News{}
	for {
		tok, err := shared.NextTag(p)
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			name := strings.ToLower(p.Name)

			if name == "publication" {
				//newsname for feed
				result, err := sp.parsePublication(p)
				if err != nil {
					return nil, err
				}
				news.Name = result.Name
				news.Language = result.Language
			} else if name == "publication_date" {
				//time for link
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				news.PublicationDate = result
			} else if name == "title" {
				//title for link
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				news.Title = result
			} else {
				p.Skip()
			}
		}
	}

	if err = p.Expect(xpp.EndTag, "news"); err != nil {
		return nil, err
	}

	return news, nil
}

func (sp *Parser) parseImage(p *xpp.XMLPullParser) (image *Image, err error) {
	if err = p.Expect(xpp.StartTag, "image"); err != nil {
		return nil, err
	}
	image = &Image{}
	for {
		tok, err := shared.NextTag(p)
		if err != nil {
			return image, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			name := strings.ToLower(p.Name)

			if name == "loc" {
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				image.Link = result
			} else {
				p.Skip()
			}
		}
	}

	if err = p.Expect(xpp.EndTag, "image"); err != nil {
		return nil, err
	}

	return image, nil
}

func (sp *Parser) parsePublication(p *xpp.XMLPullParser) (news *News, err error) {
	if err = p.Expect(xpp.StartTag, "publication"); err != nil {
		return nil, err
	}

	news = &News{}
	for {
		tok, err := shared.NextTag(p)
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			name := strings.ToLower(p.Name)

			if name == "name" {
				//newsname for feed
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				news.Name = result
			} else if name == "language" {
				//time for link
				result, err := shared.ParseText(p)
				if err != nil {
					return nil, err
				}
				news.Language = result
			} else {
				p.Skip()
			}
		}
	}

	if err = p.Expect(xpp.EndTag, "publication"); err != nil {
		return nil, err
	}

	return news, nil
}
