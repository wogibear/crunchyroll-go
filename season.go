package crunchyroll

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type Season struct {
	crunchy *Crunchyroll

	ID             string   `json:"id"`
	Title          string   `json:"title"`
	SlugTitle      string   `json:"slug_title"`
	SeriesID       string   `json:"series_id"`
	SeasonNumber   int      `json:"season_number"`
	IsComplete     bool     `json:"is_complete"`
	Description    string   `json:"description"`
	Keywords       []string `json:"keywords"`
	SeasonTags     []string `json:"season_tags"`
	IsMature       bool     `json:"is_mature"`
	MatureBlocked  bool     `json:"mature_blocked"`
	IsSubbed       bool     `json:"is_subbed"`
	IsDubbed       bool     `json:"is_dubbed"`
	IsSimulcast    bool     `json:"is_simulcast"`
	SeoTitle       string   `json:"seo_title"`
	SeoDescription string   `json:"seo_description"`

	Language LOCALE
}

// SeasonFromID returns a season by its api id
func SeasonFromID(crunchy *Crunchyroll, id string) (*Season, error) {
	resp, err := crunchy.Client.Get(fmt.Sprintf("https://beta-api.crunchyroll.com/cms/v2/%s/%s/%s/seasons/%s?locale=%s&Signature=%s&Policy=%s&Key-Pair-Id=%s",
		crunchy.Config.CountryCode,
		crunchy.Config.MaturityRating,
		crunchy.Config.Channel,
		id,
		crunchy.Locale,
		crunchy.Config.Signature,
		crunchy.Config.Policy,
		crunchy.Config.KeyPairID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var jsonBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&jsonBody)

	season := &Season{
		crunchy: crunchy,
	}
	if err := decodeMapToStruct(jsonBody, season); err != nil {
		return nil, err
	}

	return season, nil
}

// Episodes returns all episodes which are available for the season
func (s *Season) Episodes() (episodes []*Episode, err error) {
	resp, err := s.crunchy.request(fmt.Sprintf("https://beta-api.crunchyroll.com/cms/v2/%s/%s/%s/episodes?season_id=%s&locale=%s&Signature=%s&Policy=%s&Key-Pair-Id=%s",
		s.crunchy.Config.CountryCode,
		s.crunchy.Config.MaturityRating,
		s.crunchy.Config.Channel,
		s.ID,
		s.crunchy.Locale,
		s.crunchy.Config.Signature,
		s.crunchy.Config.Policy,
		s.crunchy.Config.KeyPairID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var jsonBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&jsonBody)

	for _, item := range jsonBody["items"].([]interface{}) {
		episode := &Episode{
			crunchy: s.crunchy,
		}
		if err = decodeMapToStruct(item, episode); err != nil {
			return nil, err
		}
		if episode.Playback != "" {
			streamHref := item.(map[string]interface{})["__links__"].(map[string]interface{})["streams"].(map[string]interface{})["href"].(string)
			if match := regexp.MustCompile(`(?m)^/cms/v2/\S+videos/(\w+)/streams$`).FindAllStringSubmatch(streamHref, -1); len(match) > 0 {
				episode.StreamID = match[0][1]
			}
		}
		episodes = append(episodes, episode)
	}

	return
}
