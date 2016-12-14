package gofeed

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shuyaoyimei/gofeed/atom"
	"github.com/shuyaoyimei/gofeed/rss"
	"github.com/shuyaoyimei/gofeed/sitemap"
)

// HTTPError represents an HTTP error returned by a server.
type HTTPError struct {
	StatusCode int
	Status     string
}

func (err HTTPError) Error() string {
	return fmt.Sprintf("http error: %s", err.Status)
}

// Parser is a universal feed parser that detects
// a given feed type, parsers it, and translates it
// to the universal feed type.
type Parser struct {
	AtomTranslator    Translator
	RSSTranslator     Translator
	SitemapTranslator Translator
	Client            *http.Client
	rp                *rss.Parser
	ap                *atom.Parser
	sp                *sitemap.Parser
}

// NewParser creates a universal feed parser.
func NewParser() *Parser {
	fp := Parser{
		rp: &rss.Parser{},
		ap: &atom.Parser{},
		sp: &sitemap.Parser{},
	}
	return &fp
}

// Parse parses a RSS or Atom feed into
// the universal gofeed.Feed.  It takes an
// io.Reader which should return the xml content.
func (f *Parser) Parse(feed io.Reader) (*Feed, error) {
	// Wrap the feed io.Reader in a io.TeeReader
	// so we can capture all the bytes read by the
	// DetectFeedType function and construct a new
	// reader with those bytes intact for when we
	// attempt to parse the feeds.
	var buf bytes.Buffer
	tee := io.TeeReader(feed, &buf)
	feedType := DetectFeedType(tee)

	// Glue the read bytes from the detect function
	// back into a new reader
	r := io.MultiReader(&buf, feed)

	switch feedType {
	case FeedTypeAtom:
		return f.parseAtomFeed(r)
	case FeedTypeRSS:
		return f.parseRSSFeed(r)
	case FeedTypeSitemap:
		return f.parseSitemapFeed(r)
	}

	return nil, errors.New("Failed to detect feed type")
}

// ParseURL fetches the contents of a given url and
// attempts to parse the response into the universal feed type.
func (f *Parser) ParseURL(feedURL string) (feed *Feed, err error) {
	client := f.httpClient()
	resp, err := client.Get(feedURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
	}
	defer func() {
		ce := resp.Body.Close()
		if ce != nil {
			err = ce
		}
	}()

	return f.Parse(resp.Body)
}

//ParseURLWithProxy is add proxy for pasre
func (f *Parser) ParseURLWithProxy(feedURL string, proxyURL string, proxyName string, proxyPasswd string) (feed *Feed, err error) {
	client := f.httpClientWithProxy(proxyURL)
	req, _ := http.NewRequest("GET", feedURL, nil)
	basePas := base64.StdEncoding.EncodeToString([]byte(proxyName + ":" + proxyPasswd))
	req.Header.Set("Proxy-Authorization", "Basic "+basePas)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
	}
	defer func() {
		ce := resp.Body.Close()
		if ce != nil {
			err = ce
		}
	}()
	return f.Parse(resp.Body)
}

// ParseString parses a feed XML string and into the
// universal feed type.
func (f *Parser) ParseString(feed string) (*Feed, error) {
	return f.Parse(strings.NewReader(feed))
}

func (f *Parser) parseAtomFeed(feed io.Reader) (*Feed, error) {
	af, err := f.ap.Parse(feed)
	if err != nil {
		return nil, err
	}
	return f.atomTrans().Translate(af)
}

func (f *Parser) parseRSSFeed(feed io.Reader) (*Feed, error) {
	rf, err := f.rp.Parse(feed)
	if err != nil {
		return nil, err
	}

	return f.rssTrans().Translate(rf)
}

func (f *Parser) parseSitemapFeed(feed io.Reader) (*Feed, error) {
	sf, err := f.sp.Parse(feed)
	if err != nil {
		return nil, err
	}

	return f.sitemapTrans().Translate(sf)
}

func (f *Parser) atomTrans() Translator {
	if f.AtomTranslator != nil {
		return f.AtomTranslator
	}
	f.AtomTranslator = &DefaultAtomTranslator{}
	return f.AtomTranslator
}

func (f *Parser) rssTrans() Translator {
	if f.RSSTranslator != nil {
		return f.RSSTranslator
	}
	f.RSSTranslator = &DefaultRSSTranslator{}
	return f.RSSTranslator
}

func (f *Parser) sitemapTrans() Translator {
	if f.SitemapTranslator != nil {
		return f.SitemapTranslator
	}
	f.SitemapTranslator = &DefaultSitemapTranslator{}
	return f.SitemapTranslator
}

func (f *Parser) httpClient() *http.Client {
	if f.Client != nil {
		return f.Client
	}
	timeout := time.Duration(15 * time.Second)

	f.Client = &http.Client{Timeout: timeout}
	return f.Client
}

func (f *Parser) httpClientWithProxy(uRLProxy string) *http.Client {
	if f.Client != nil {
		return f.Client
	}
	//uRLProxy must be xxx.xxx.xxx.xxx:prot
	timeout := time.Duration(15 * time.Second)
	urlProxys := &url.URL{Host: uRLProxy}
	// urlProxys, _ := url.Parse(uRLProxy)
	f.Client = &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			Proxy: http.ProxyURL(urlProxys),
		},
		Timeout: timeout,
	}
	return f.Client
}
