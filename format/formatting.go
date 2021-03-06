package format

import "encoding/json"
import "encoding/xml"
import "errors"
import "github.com/peaberberian/OscarGoGo/config"

func ParseFeed(raw []byte, web config.Website) (FeedFormat, error) {
	var feedRes FeedFormat

	switch web.FeedFormat {

	// parse RSS Feeds
	case "rss":
		var xmlBody rssFormat
		err := xml.Unmarshal(raw, &xmlBody)
		if err != nil {
			return FeedFormat{}, err
		} else {
			feedRes = parseRss(xmlBody, web)
		}

	// parse Atom feeds
	case "atom":
		var xmlBody atomFormat
		err := xml.Unmarshal(raw, &xmlBody)
		if err != nil {
			return FeedFormat{}, err
		} else {
			feedRes = parseAtom(xmlBody, web)
		}

	// Try to autodetect Feed type (duck-typing)
	default:
		var rssRaw rssFormat
		var atomRaw atomFormat
		errRss := xml.Unmarshal(raw, &rssRaw)
		errAtom := xml.Unmarshal(raw, &atomRaw)

		if errRss == nil && (len(rssRaw.Channels.Items) > 0 ||
			rssRaw.Channels.Title != "") {
			ret := parseRss(rssRaw, web)
			return ret, nil
		}

		if errAtom == nil && (len(atomRaw.Entries) > 0 ||
			atomRaw.Title != "") {
			ret := parseAtom(atomRaw, web)
			return ret, nil
		}

		if errRss != nil {
			return FeedFormat{}, errRss
		}
		if errAtom != nil {
			return FeedFormat{}, errAtom
		}
		return FeedFormat{}, errors.New("Could not detect your feed format")
	}
	return feedRes, nil
}

// Convert an rssFormat to a FeedFormat
func parseRss(rssMap rssFormat, web config.Website) FeedFormat {
	var feedTime = parseRssTime(rssMap.Channels.PubDate)
	var feed = FeedFormat{
		Id:          web.Id,
		Title:       rssMap.Channels.Title,
		Link:        web.FeedLink,
		Description: rssMap.Channels.Description,
		UpdateDate:  feedTime,
	}
	for _, item := range rssMap.Channels.Items {
		var date = parseRssTime(item.PubDate)

		feed.Entries = append(feed.Entries, feedEntry{
			Title:        item.Title,
			Link:         item.Link,
			Description:  item.Description,
			CreationDate: date,
			UpdateDate:   date,
		})
	}
	return feed
}

// Convert an atomFormat to a FeedFormat
func parseAtom(atomMap atomFormat, web config.Website) FeedFormat {
	var feedTime = parseAtomTime(atomMap.Updated)
	var feed = FeedFormat{
		Id:          web.Id,
		Title:       atomMap.Title,
		Link:        web.FeedLink,
		Description: atomMap.Subtitle,
		UpdateDate:  feedTime,
	}
	for _, item := range atomMap.Entries {
		var date = parseAtomTime(item.Updated)

		var description string
		if item.Summary != "" {
			description = item.Summary
		} else {
			description = item.Content
		}
		feed.Entries = append(feed.Entries, feedEntry{
			Title:        item.Title,
			Link:         item.Links[0].Key,
			Description:  description,
			CreationDate: date,
			UpdateDate:   date,
		})
	}
	return feed
}

func ConvertFeedsToJson(feeds []FeedFormat) ([]byte, error) {
	var jsobjs = []jsonFormat{}
	for _, feed := range feeds {
		var jsobj = jsonFormat{
			Id:   feed.Id,
			Name: feed.Title,
			Link: feed.Link,
		}

		for _, entry := range feed.Entries {
			jsobj.Items = append(jsobj.Items, jsonItem{
				Title:        entry.Title,
				Link:         entry.Link,
				Description:  entry.Description,
				CreationDate: timeToString(entry.CreationDate),
			})
		}
		jsobjs = append(jsobjs, jsobj)
	}
	res, err := json.Marshal(jsobjs)
	if err != nil {
		return []byte{}, err
	}
	return res, nil
}

func ConvertWebsitesToJson(webs []config.Website) ([]byte, error) {
	var websJson []websiteJSON
	for _, web := range webs {
		websJson = append(websJson, websiteJSON{
			Id:          web.Id,
			Description: web.Description,
			FeedLink:    web.FeedLink,
			FeedName:    web.FeedName,
			FeedFormat:  web.FeedFormat,
			SiteLink:    web.SiteLink,
			SiteName:    web.SiteName,
		})
	}
	ret, err := json.Marshal(websJson)
	if err != nil {
		return []byte{}, err
	}
	return ret, nil
}

// needed or not?
func AutoDetectFeedFormat(raw []byte) (string, error) {
	var rssRaw rssFormat
	var atomRaw atomFormat
	errRss := xml.Unmarshal(raw, &rssRaw)
	errAtom := xml.Unmarshal(raw, &atomRaw)
	if errRss == nil && (len(rssRaw.Channels.Items) > 0 ||
		rssRaw.Channels.Title != "") {
		return "rss", nil
	}

	if errAtom == nil && (len(atomRaw.Entries) > 0 ||
		atomRaw.Title != "") {
		return "atom", nil
	}

	if errRss != nil {
		return "", errRss
	}
	if errAtom != nil {
		return "", errAtom
	}

	return "", errors.New("Could not detect your feed format")
}
