package extrainfo


import (

	"github.com/badoux/goscraper"
)

//GetInfoWebsite get logo and title from a website
func GetInfoWebsite(url string) (string, string) {
	url = "http://" + url
	var icon string
	var title string
	s, err := goscraper.Scrape(url, 5)
	if err != nil {
		icon = "No icon"
		title = "No title"
		return icon, title

	}
	if s.Preview.Icon == "" {
		icon = "No icon"
	} else {
		icon = s.Preview.Icon
	}

	if s.Preview.Title == "" {
		title = "No Title"
	} else {
		title = s.Preview.Title
	}
	return icon, title
}

