package www

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"strings"
	"time"
)

type Sitemap struct {
	URLSet []SitemapURL `xml:"urlset"`
}

func NewSitemap(baseURL string, fsys fs.FS) (*Sitemap, error) {
	// Get HTML page paths from file system
	pages := []string{}
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Add path only if page is HTML and ignore 404 page
		if !strings.HasSuffix(path, ".html") || strings.HasSuffix(path, "404.html") {
			return nil
		}
		pages = append(pages, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get HTML pages from FS: %w", err)
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("no HTML page in FS")
	}

	// Prepend base URL to each page path
	for i, page := range pages {
		pages[i] = baseURL + page
	}

	// Add URLs to set
	out := &Sitemap{}
	for _, url := range pages {
		out.URLSet = append(out.URLSet, SitemapURL{
			Location:   url,
			LastMod:    time.Now(),
			ChangeFreq: "monthly",
			Priority:   0.9,
		})
	}

	return out, nil
}

type SitemapURL struct {
	XMLName    xml.Name  `xml:"url"`
	Location   string    `xml:"location"`
	LastMod    time.Time `xml:"lastmod"`
	ChangeFreq string    `xml:"changefreq,omitempty"`
	Priority   float32   `xml:"priority,omitempty"`
}

func (s *Sitemap) String() (string, error) {
	out := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
	out += "<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n"

	urlsetContent, err := xml.Marshal(s.URLSet)
	if err != nil {
		return "", err
	}
	out += "\t" + string(urlsetContent) + "\n"
	out += "</urlset>"

	return out, nil
}
