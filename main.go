package main

import (
	"bytes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"regexp"
	"strings"
	"time"

	"go.senan.xyz/taglib"
	"golang.org/x/crypto/blowfish"
)

type Playlist struct {
	ID             int     `json:"id"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Duration       int     `json:"duration"`
	Public         bool    `json:"public"`
	IsLovedTrack bool    `json:"is_loved_track"`
	Collaborative  bool    `json:"collaborative"`
	NbTracks      int     `json:"nb_tracks"`
	Fans           int     `json:"fans"`
	Link           string  `json:"link"`
	Share          string  `json:"share"`
	Picture        string  `json:"picture"`
	PictureSmall  string  `json:"picture_small"`
	PictureMedium string  `json:"picture_medium"`
	PictureBig    string  `json:"picture_big"`
	PictureXL     string  `json:"picture_xl"`
	Checksum       string  `json:"checksum"`
	Tracklist      string  `json:"tracklist"`
	CreationDate  string  `json:"creation_date"`
	AddDate       string  `json:"add_date"`
	ModDate       string  `json:"mod_date"`
	MD5Image      string  `json:"md5_image"`
	PictureType   string  `json:"picture_type"`
	Creator        Creator `json:"creator"`
	Tracks         Tracks  `json:"tracks"`
}

type Creator struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Tracklist string `json:"tracklist"`
	Type      string `json:"type"`
}

type Tracks struct {
	Data     []PlaylistData `json:"data"`
	Checksum string         `json:"checksum"`
}

type PlaylistData struct {
	ID                      int    `json:"id"`
	Readable                bool   `json:"readable"`
	Title                   string `json:"title"`
	TitleShort             string `json:"title_short"`
	TitleVersion           string `json:"title_version"`
	ISRC                    string `json:"isrc"`
	Link                    string `json:"link"`
	Duration                int    `json:"duration"`
	Rank                    int    `json:"rank"`
	ExplicitLyrics         bool   `json:"explicit_lyrics"`
	ExplicitContentLyrics int    `json:"explicit_content_lyrics"`
	ExplicitContentCover  int    `json:"explicit_content_cover"`
	Preview                 string `json:"preview"`
	MD5Image               string `json:"md5_image"`
	TimeAdd                int    `json:"time_add"`

	Artist struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Link      string `json:"link"`
		Tracklist string `json:"tracklist"`
		Type      string `json:"type"`
	}

	Album Album  `json:"album"`
	Type  string `json:"type"`
}

type FetchedDataPlaylist struct {
	PlaylistName string
	Tracks       []string
}

type Albums struct {
	Data     []Album `json:"data"`
	Checksum string  `json:"checksum"`
	Total    int     `json:"total"`
	Next     string  `json:"next"`
}

type Album struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Link         string `json:"link"`
	Upc          string `json:"upc"`
	Cover        string `json:"cover"`
	CoverSmall  string `json:"cover_small"`
	CoverMedium string `json:"cover_medium"`
	CoverBig    string `json:"cover_big"`
	CoverXL     string `json:"cover_xl"`
	MD5Image    string `json:"md5_image"`
	Genres       struct {
		Data []GenreData `json:"data"`
	} `json:"genres"`
	Label     string `json:"label"`
	Tracklist string `json:"tracklist"`
	Type      string `json:"type"`
	Artist    Artist `json:"artist"`
}

type GenreData struct {
	Name string `json:"name"`
}
type Artist struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Link           string `json:"link"`  // new parameters
	Share          string `json:"share"` // new parameters
	Picture        string `json:"picture"`
	PictureSmall  string `json:"picture_small"`
	PictureMedium string `json:"picture_medium"`
	PictureBig    string `json:"picture_big"`
	PictureXL     string `json:"picture_xl"`
	Radio          bool   `json:"radio"` // new parameters
	Tracklist      string `json:"tracklist"`
	Type           string `json:"type"`
}
type FavoriteTracks struct {
	ID                      int       `json:"id"`
	Readable                bool      `json:"readable"`
	Title                   string    `json:"title"`
	Link                    string    `json:"link"`
	Duration                int       `json:"duration"`
	Rank                    int       `json:"rank"`
	ExplicitLyrics         bool      `json:"explicit_lyrics"`
	ExplicitContentLyrics int       `json:"explicit_content_lyrics"`
	ExplicitContentCover  int       `json:"explicit_content_cover"`
	MD5Image               string    `json:"md5_image"`
	TimeAdd                string	  `json:"time_add"`
	Album                   Album     `json:"album"`
	Artist                  Artist    `json:"artist"`
	Type                    string    `json:"type"`
}

type User struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Link           string `json:"link"`
	Picture        string `json:"picture"`
	PictureSmall  string `json:"picture_small"`
	PictureMedium string `json:"picture_medium"`
	PictureBig    string `json:"picture_big"`
	PictureXL     string `json:"picture_xl"`
	Country        string `json:"country"`
	Tracklist      string `json:"tracklist"`
	Type           string `json:"type"`
}

type DeezerError struct {
	Err struct {
		Type    string `json:"type"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
}

type Track struct {
	ID                      int            `json:"id"`
	Readable                bool           `json:"readable"`
	Title                   string         `json:"title"`
	TitleShort             string         `json:"title_short"`
	TitleVersion           string         `json:"title_version"`
	ISRC                    string         `json:"isrc"`
	Link                    string         `json:"link"`
	Share                   string         `json:"share"`
	Duration                int            `json:"duration"`
	TrackPosition          int            `json:"track_position"`
	DiskNumber             int            `json:"disk_number"`
	Rank                    int            `json:"rank"`
	ReleaseDate            string         `json:"releaseDate"`
	ExplicitLyrics         bool           `json:"explicit_lyrics"`
	ExplicitContentLyrics 	int            `json:"explicit_content_lyrics"`
	ExplicitContentCover  int            `json:"explicit_content_cover"`
	Preview                 string         `json:"preview"`
	BPM                     float64        `json:"bpm"`
	Gain                    float64        `json:"gain"`
	AvailableCountries      []string       `json:"available_countries"`
	Contributors            []Contributors `json:"contributors"`
	MD5Image                string         `json:"md5_image"`
	TrackToken              string         `json:"trackToken"`
	Artist                  Artist         `json:"artist"`
	Album                   Album          `json:"album"`
}

type Contributors struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Link           string `json:"link"`
	Share          string `json:"share"`
	Picture        string `json:"picture"`
	PictureSmall  string `json:"picture_small"`
	PictureMedium string `json:"picture_medium"`
	PictureBig    string `json:"picture_big"`
	PictureXL     string `json:"picture_xl"`
	Radio          bool   `json:"radio"`
	Tracklist      string `json:"tracklist"`
	Type           string `json:"type"`
	Role           string `json:"role"`
}

type GatewayResponse struct {
	DeezerError []DeezerError `json:"error"`
	Results     struct {
		User struct {
			Options struct {
				LicenseToken string `json:"licenseToken"`
			}
		}
	}
}

var client *http.Client


func initClient() {
	jar, _ := cookiejar.New(nil)
	client = &http.Client{
		Jar: jar,
	}
}

func getJSON(url string, v any) error {
	response, err := http.Get(url)

	if err != nil {
		return fmt.Errorf("get %s: %w", url, err)
	} else if response.StatusCode != http.StatusOK {
		defer response.Body.Close()
		return fmt.Errorf("get %s: unexpected status %s", url, response.Status)
	}

	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(v)
	if err != nil {
		return fmt.Errorf("decode %s: %w", url, err)
	}
	return nil
}

func fetchPlaylist(id int) (FetchedDataPlaylist, error) {
	url := fmt.Sprintf("https://api.deezer.com/playlist/%d", id)
	var result Playlist
	err := getJSON(url, &result)
	if err != nil {
		return FetchedDataPlaylist{}, err
	}

	var tracks []string
	for _, track := range result.Tracks.Data {
		tracks = append(tracks, fmt.Sprintf("%s - %s", track.Artist.Name, track.Title))
	}
	var data FetchedDataPlaylist
	data.PlaylistName = result.Title
	data.Tracks = tracks
	return data, nil
}

func createPlaylistFile(data FetchedDataPlaylist) error {
	re := regexp.MustCompile(`\s`)
	fileName := strings.ToLower(re.ReplaceAllString(data.PlaylistName, "_"))
	fileName = strings.ReplaceAll(fileName, "/", "_")

	musicListFile, err := os.Create(fileName + ".m3u")
	if err != nil {
		return err
	}

	musicList := data.Tracks
	for _, track := range musicList {
		musicListFile.WriteString(track + "\n")
	}

	fmt.Printf("Tracklist has been written in %s.txt", fileName)
	return nil
}

func fetchAlbums(uid int) ([]string, error) {
	url := fmt.Sprintf("https://api.deezer.com/user/%d/albums", uid)

	var result Albums
	getJSON(url, &result)
	var albums []string
	for url != "" {
		var result Albums
		err := getJSON(url, &result)
		if err != nil {
			return []string{}, err
		}

		url = result.Next
		for _, album := range result.Data {
			albums = append(albums, fmt.Sprintf("%s - %s", album.Artist.Name, album.Title))
		}
	}

	return albums, nil
}

func getUsername(uid int) (string, error) {
	url := fmt.Sprintf("https://api.deezer.com/user/%d", uid)
	var result User
	err := getJSON(url, &result)
	if err != nil {
		return "", err
	}
	return result.Name, nil
}

func createAlbumFile(data []string) {}

func fetchFavoriteTracks(uid int) ([]string, error) {
	url := fmt.Sprintf("https://api.deezer.com/user/%d/tracks", uid)
	var result Albums
	err := getJSON(url, &result)
	if err != nil {
		return []string{}, err
	}

	var tracks []string
	for url != "" {
		var result Albums
		err := getJSON(url, &result)
		if err != nil {
			return []string{}, err
		}

		url = result.Next
		for _, albumTrack := range result.Data {
			track, err := getTrackInfo(fmt.Sprintf("%d", albumTrack.ID))
			if err != nil {
				return []string{}, err
			}

			trackName := formatTrackName(track)
			tracks = append(tracks, trackName)
		}
	}
	return tracks, nil
}

func calcBlowfishKey(trackID string) []byte {
	secret := []byte("g4el58wc0zvf9na1")
	hash := md5.Sum([]byte(trackID))
	hashHex := hex.EncodeToString(hash[:])

	key := make([]byte, len(secret))
	copy(key, secret)

	for i := range len(hash) {
		key[i] = key[i] ^ hashHex[i] ^ hashHex[i+16]
	}

	return key
}

func decryptChunk(data []byte, key []byte) ([]byte, error) {
	blowfishIV := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}

	bf, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypted := make([]byte, len(data))

	cipher.NewCBCDecrypter(bf, blowfishIV).CryptBlocks(decrypted, data)
	return decrypted, nil
}

func decryptFile(enc io.Reader, key []byte, dec io.Writer) error {
	const blockSize = 2048

	buf := make([]byte, blockSize)
	for i := 0; ; i++ {
		n, err := io.ReadFull(enc, buf)
		if err == io.EOF {
			break
		}

		if err != nil && err != io.ErrUnexpectedEOF {
			return err
		}

		data := buf[:n]
		isEncrypted := i%3 == 0
		isWholeBlock := n == blockSize

		if isEncrypted && isWholeBlock {
			data, err = decryptChunk(data, key)
			if err != nil {
				return err
			}
		}

		if _, werr := dec.Write(data); werr != nil {
			return werr
		}

		if err == io.ErrUnexpectedEOF {
			break
		}

	}
	return nil
}

func getLicenseToken(arl string) (string, error) {
	ajaxURL := "https://www.deezer.com/ajax/gw-light.php?method=deezer.getUserData&input=3&api_version=1.0&api_token="
	req, err := http.NewRequest("GET", ajaxURL, http.NoBody)

	if err != nil {
		return "", err
	}

	req.Header.Set("Cookie", "arl="+arl)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var result GatewayResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	licenseToken := result.Results.User.Options.LicenseToken
	return licenseToken, nil
}

func getTrackToken(arl string, trackID string) (string, error) {
	apiEndpoint := fmt.Sprintf("https://api.deezer.com/track/%s", trackID)
	req, err := http.NewRequest("GET", apiEndpoint, http.NoBody)
	if err != nil {
		return "", nil
	}

	req.Header.Set("Cookie", "arl="+arl+"; comeback=1")
	resp, err := client.Do(req)
	if err != nil {
		return "", nil
	}

	defer resp.Body.Close()

	var result Track
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	trackToken := result.TrackToken
	return trackToken, nil
}

func createMediaPayload(licenseToken, trackToken string) ([]byte, error) {
	type formats struct {
		Cipher string `json:"cipher"`
		Format string `json:"format"`
	}
	type media struct {
		Type    string    `json:"type"`
		Formats []formats `json:"formats"`
	}

	type payload struct {
		LicenseToken string   `json:"licenseToken"`
		Media        []media  `json:"media"`
		TrackTokens  []string `json:"trackTokens"`
	}

	pl := payload{
		LicenseToken: licenseToken,
		Media: []media{
			{
				Type: "FULL",
				Formats: []formats{
					{Cipher: "BF_CBC_STRIPE", Format: "MP3_128"},
				},
			},
		},
		TrackTokens: []string{trackToken},
	}
	result, err := json.Marshal(pl)
	if err != nil {
		return []byte{}, err
	}
	return result, nil
}

func getTrackURL(licenseToken, trackToken string) (string, error) {
	type media struct {
		MediaType string `json:"media_type"`
		Cipher    struct {
			Type string `json:"type"`
		} `json:"cipher"`
		Format  string `json:"format"`
		Sources []struct {
			URL string `json:"url"`
		} `json:"sources"`
	}

	type data struct {
		Media []media `json:"media"`
	}

	type getURLResponse struct {
		Data []data `json:"data"`
	}

	getURLEndpoint := "https://media.deezer.com/v1/get_url"
	payload, err := createMediaPayload(licenseToken, trackToken)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", getURLEndpoint, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	var result getURLResponse
	err = json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Data) == 0 {
		return "", fmt.Errorf("No data from response")
	}

	for _, dataItem := range result.Data {
		if len(dataItem.Media) == 0 {
			continue
		}
		for _, mediaItem := range dataItem.Media {
			if len(mediaItem.Sources) == 0 {
				continue
			}
			for _, source := range mediaItem.Sources {
				if len(source.URL) == 0 {
					continue
				}
				url := source.URL
				return url, nil
			}
		}
	}
	return "", err
}

func formatTrackName(trackInfo Track) string {
	var artistsBuilder strings.Builder

	for i, contributor := range trackInfo.Contributors {
		artistsBuilder.WriteString(contributor.Name)
		if i < len(trackInfo.Contributors)-1 {
			artistsBuilder.WriteString(", ")
		}
	}

	artists := artistsBuilder.String()

	fileName := fmt.Sprintf("%s - %s", artists, trackInfo.Title)
	fileName = strings.ReplaceAll(fileName, "/", "_")
	return fileName
}

func downloadTrack(url string, trackInfo Track) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var enc bytes.Buffer

	fileName := formatTrackName(trackInfo) + ".mp3"

	_, err = io.Copy(&enc, resp.Body)

	trackID := fmt.Sprintf("%d", trackInfo.ID)
	key := calcBlowfishKey(trackID)
	if err != nil {
		return "", err
	}

	dec, err := os.Create(fileName)
	if err != nil {
		return "", err
	}

	decryptFile(&enc, key, dec)

	return fileName, nil
}

func addMetadata(fileName string, trackInfo Track) error {
	var artists []string
	for _, contributor := range trackInfo.Contributors {
		artists = append(artists, contributor.Name)
	}

	albumURL := fmt.Sprintf("https://api.deezer.com/album/%d", trackInfo.Album.ID)
	var albumInfo Album

	err := getJSON(albumURL, &albumInfo)
	if err != nil {
		return err
	}

	var genre string
	for _, g := range albumInfo.Genres.Data {
		genre = g.Name
	}

	const dateFormat = "2006-01-02"
	releaseDate, err := time.Parse(dateFormat, trackInfo.ReleaseDate)
	if err != nil {
		return err
	}

	err = taglib.WriteTags(fileName, map[string][]string{
		taglib.Artist:      artists,
		taglib.Title:       {trackInfo.Title},
		taglib.Album:       {trackInfo.Album.Title},
		taglib.AlbumArtist: {trackInfo.Album.Artist.Name},
		taglib.TrackNumber: {fmt.Sprintf("%d", trackInfo.TrackPosition)},
		taglib.ISRC:        {trackInfo.ISRC},
		taglib.BPM:         {fmt.Sprintf("%f", trackInfo.BPM)},
		taglib.Genre:       {genre},
		taglib.Label:       {albumInfo.Label},
		"Year":             {fmt.Sprintf("%d", releaseDate.Year())},
		"Gain":             {fmt.Sprintf("%f", trackInfo.Gain)},
	}, 0)

	coverURL := trackInfo.Album.CoverBig

	resp, err := http.Get(coverURL)
	if err != nil {
		return err
	}

	var cover bytes.Buffer

	_, err = io.Copy(&cover, resp.Body)
	err = taglib.WriteImageOptions(
		fileName,
		cover.Bytes(), 0,
		"Front Cover",
		"Front artwork",
		"image/jpeg")

	if err != nil {
		return err
	}

	return nil
}

func getTrackInfo(trackID string) (Track, error) {
	apiEndpoint := fmt.Sprintf("https://api.deezer.com/track/%s", trackID)
	var track Track
	err := getJSON(apiEndpoint, &track)
	if err != nil {
		return Track{}, err
	}

	return track, nil
}

func main() {
	initClient()

	arl := os.Getenv("arl")
	//userID := os.Getenv("deezer_uid")
	//songID := "67929161" // track with a single artsit
	songID := "67929169" // track with multiple artists

	licenseToken, err := getLicenseToken(arl)
	if err != nil {
		log.Fatal(err)
	}

	trackToken, err := getTrackToken(arl, songID)
	if err != nil {
		log.Fatal(err)
	}

	trackURL, err := getTrackURL(licenseToken, trackToken)
	if err != nil {
		log.Fatal(err)
	}

	trackInfo, err := getTrackInfo(songID)
	if err != nil {
		log.Fatal(err)
	}

	fileName, err := downloadTrack(trackURL, trackInfo)
	if err != nil {
		log.Fatal(err)
	}

	err = addMetadata(fileName, trackInfo)
	if err != nil {
		log.Fatal(err)
	}

	//data, err := fetchPlaylist(playlistID)
	//if err != nil {
	//	log.Fatal(err)
	//}

	////albums := fetchAlbums(userID)
	//for _, track := range data.Tracks {
	//	fmt.Println(track)time
	//}

	//favoriteTracks, err := fetchFavoriteTracks(userID)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//for _, track := range favoriteTracks {
	//	fmt.Println(track)
	//}

}
