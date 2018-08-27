package video_providers

import (
	"fmt"
	"strconv"

	"github.com/tellytv/telly/internal/m3uplus"
	"github.com/tellytv/telly/utils"
)

type M3U struct {
	BaseConfig Configuration

	Playlist           *m3uplus.Playlist
	channels           map[int]Channel
	categoriesStrCheck []string
	categories         []Category
	seenFormats        []string
}

func newM3U(config *Configuration) (VideoProvider, error) {
	m3u := &M3U{BaseConfig: *config}

	if loadErr := m3u.Refresh(); loadErr != nil {
		return nil, loadErr
	}

	return m3u, nil
}

func (m *M3U) Name() string {
	return "M3U"
}

func (m *M3U) Categories() ([]Category, error) {
	return m.categories, nil
}

func (m *M3U) Formats() ([]string, error) {
	return m.seenFormats, nil
}

func (m *M3U) Channels() ([]Channel, error) {
	outputChannels := make([]Channel, 0)
	for _, channel := range m.channels {
		outputChannels = append(outputChannels, channel)
	}
	return outputChannels, nil
}

func (m *M3U) StreamURL(streamID int, wantedFormat string) (string, error) {
	if val, ok := m.channels[streamID]; ok {
		return val.streamUrl, nil
	}
	return "", fmt.Errorf("that channel id (%d) does not exist in the video source lineup", streamID)
}

func (m *M3U) Refresh() error {
	playlist, m3uErr := utils.GetM3U(m.BaseConfig.M3UURL, false)
	if m3uErr != nil {
		return fmt.Errorf("error when reading m3u: %s", m3uErr)
	}
	m.Playlist = playlist

	for _, track := range playlist.Tracks {
		streamURL := streamNumberRegex(track.URI, -1)[0]

		channelID, channelIDErr := strconv.Atoi(streamURL[1])
		if channelIDErr != nil {
			return fmt.Errorf("error when extracting channel id from m3u track: %s", channelIDErr)
		}

		if !utils.Contains(m.seenFormats, streamURL[2]) {
			m.seenFormats = append(m.seenFormats, streamURL[2])
		}

		nameVal := track.Name

		if val, ok := track.Tags["tvg-name"]; ok {
			nameVal = val
		}

		if m.BaseConfig.NameKey != "" {
			if val, ok := track.Tags[m.BaseConfig.NameKey]; ok {
				nameVal = val
			}
		}

		logoVal := track.Tags["tvg-logo"]
		if m.BaseConfig.LogoKey != "" {
			if val, ok := track.Tags[m.BaseConfig.LogoKey]; ok {
				logoVal = val
			}
		}

		categoryVal := track.Tags["group-title"]
		if m.BaseConfig.CategoryKey != "" {
			if val, ok := track.Tags[m.BaseConfig.CategoryKey]; ok {
				categoryVal = val
			}
		}

		if !utils.Contains(m.categoriesStrCheck, categoryVal) {
			m.categoriesStrCheck = append(m.categoriesStrCheck, categoryVal)
			m.categories = append(m.categories, Category{
				Name: categoryVal,
				Type: "live",
			})
		}

		epgIDVal := track.Tags["tvg-id"]
		if m.BaseConfig.EPGIDKey != "" {
			if val, ok := track.Tags[m.BaseConfig.EPGIDKey]; ok {
				epgIDVal = val
			}
		}

		m.channels[channelID] = Channel{
			Name:     nameVal,
			StreamID: channelID,
			Logo:     logoVal,
			Type:     ChannelType(LiveStream),
			Category: categoryVal,
			EPGID:    epgIDVal,

			streamUrl: track.URI,
		}
	}

	return nil
}

func (m *M3U) Configuration() Configuration {
	return m.BaseConfig
}