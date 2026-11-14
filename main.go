package main

import (
	"bytes"
	"context"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	_ "strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/schollz/progressbar/v3"
	"go.senan.xyz/taglib"
	"golang.org/x/crypto/blowfish"
	"golang.org/x/sync/errgroup"
)

type Playlist struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Duration      int     `json:"duration"`
	Public        bool    `json:"public"`
	IsLovedTrack  bool    `json:"is_loved_track"`
	Collaborative bool    `json:"collaborative"`
	NbTracks      int     `json:"nb_tracks"`
	Fans          int     `json:"fans"`
	Link          string  `json:"link"`
	Share         string  `json:"share"`
	Picture       string  `json:"picture"`
	PictureSmall  string  `json:"picture_small"`
	PictureMedium string  `json:"picture_medium"`
	PictureBig    string  `json:"picture_big"`
	PictureXL     string  `json:"picture_xl"`
	Checksum      string  `json:"checksum"`
	Tracklist     string  `json:"tracklist"`
	CreationDate  string  `json:"creation_date"`
	AddDate       string  `json:"add_date"`
	ModDate       string  `json:"mod_date"`
	MD5Image      string  `json:"md5_image"`
	PictureType   string  `json:"picture_type"`
	ReleaseDate   string
	ExplicitLyrics	bool
	Creator       struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Tracklist string `json:"tracklist"`
		Type      string `json:"type"`
	} `json:"creator"`
	User       	struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Tracklist string `json:"tracklist"`
		Type      string `json:"type"`
	} `json:"user"`
	Tracks        struct {
		Data []Track `json:"data"`
	} `json:"tracks"`
}

type PlaylistData struct {
	ID                    int    `json:"id"`
	Readable              bool   `json:"readable"`
	Title                 string `json:"title"`
	TitleShort            string `json:"title_short"`
	TitleVersion          string `json:"title_version"`
	ISRC                  string `json:"isrc"`
	Link                  string `json:"link"`
	Duration              int    `json:"duration"`
	Rank                  int    `json:"rank"`
	ExplicitLyrics        bool   `json:"explicit_lyrics"`
	ExplicitContentLyrics int    `json:"explicit_content_lyrics"`
	ExplicitContentCover  int    `json:"explicit_content_cover"`
	Preview               string `json:"preview"`
	MD5Image              string `json:"md5_image"`
	TimeAdd               int    `json:"time_add"`

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
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	Upc         string `json:"upc"`
	NbTracks	int		`json:"nb_tracks"`
	Cover       string `json:"cover"`
	CoverSmall  string `json:"cover_small"`
	CoverMedium string `json:"cover_medium"`
	CoverBig    string `json:"cover_big"`
	CoverXL     string `json:"cover_xl"`
	MD5Image    string `json:"md5_image"`
	ReleaseDate           string        `json:"release_date"`
	ExplicitLyrics        bool   `json:"explicit_lyrics"`
	Genres      struct {
		Data []GenreData `json:"data"`
	} `json:"genres"`
	Label     string `json:"label"`
	Duration     int `json:"duration"`
	Tracklist string `json:"tracklist"`
	Type      string `json:"type"`
	Tracks    struct {
		Data []Track `json:"data"`
	}
	Artist Artist `json:"artist"`
}

type GenreData struct {
	Name string `json:"name"`
}
type Artist struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Link          string `json:"link"`  // new parameters
	Share         string `json:"share"` // new parameters
	Picture       string `json:"picture"`
	PictureSmall  string `json:"picture_small"`
	PictureMedium string `json:"picture_medium"`
	PictureBig    string `json:"picture_big"`
	PictureXL     string `json:"picture_xl"`
	Radio         bool   `json:"radio"` // new parameters
	Tracklist     string `json:"tracklist"`
	Type          string `json:"type"`
}

type FavoriteTracks struct {
	ID                    int    `json:"id"`
	Readable              bool   `json:"readable"`
	Title                 string `json:"title"`
	Link                  string `json:"link"`
	Duration              int    `json:"duration"`
	Rank                  int    `json:"rank"`
	ExplicitLyrics        bool   `json:"explicit_lyrics"`
	ExplicitContentLyrics int    `json:"explicit_content_lyrics"`
	ExplicitContentCover  int    `json:"explicit_content_cover"`
	MD5Image              string `json:"md5_image"`
	TimeAdd               string `json:"time_add"`
	Album                 Album  `json:"album"`
	Artist                Artist `json:"artist"`
	Type                  string `json:"type"`
}

type User struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Link          string `json:"link"`
	Picture       string `json:"picture"`
	PictureSmall  string `json:"picture_small"`
	PictureMedium string `json:"picture_medium"`
	PictureBig    string `json:"picture_big"`
	PictureXL     string `json:"picture_xl"`
	Country       string `json:"country"`
	Tracklist     string `json:"tracklist"`
	Type          string `json:"type"`
}

type DeezerError struct {
	Err struct {
		Type    string `json:"type"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
}

type Track struct {
	ID                    int           `json:"id"`
	Readable              bool          `json:"readable"`
	Title                 string        `json:"title"`
	TitleShort            string        `json:"title_short"`
	TitleVersion          string        `json:"title_version"`
	ISRC                  string        `json:"isrc"`
	Link                  string        `json:"link"`
	Share                 string        `json:"share"`
	Duration              int           `json:"duration"`
	TrackPosition         int           `json:"track_position"`
	DiskNumber            int           `json:"disk_number"`
	Rank                  int           `json:"rank"`
	ReleaseDate           string        `json:"release_date"`
	ExplicitLyrics        bool          `json:"explicit_lyrics"`
	ExplicitContentLyrics int           `json:"explicit_content_lyrics"`
	ExplicitContentCover  int           `json:"explicit_content_cover"`
	Preview               string        `json:"preview"`
	BPM                   float64       `json:"bpm"`
	Gain                  float64       `json:"gain"`
	AvailableCountries    []string      `json:"available_countries"`
	Contributors          []Contributor `json:"contributors"`
	MD5Image              string        `json:"md5_image"`
	TrackToken            string        `json:"track_token"`
	Artist                Artist        `json:"artist"`
	Album                 Album         `json:"album"`
}

type Contributor struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Link          string `json:"link"`
	Share         string `json:"share"`
	Picture       string `json:"picture"`
	PictureSmall  string `json:"picture_small"`
	PictureMedium string `json:"picture_medium"`
	PictureBig    string `json:"picture_big"`
	PictureXL     string `json:"picture_xl"`
	Radio         bool   `json:"radio"`
	Tracklist     string `json:"tracklist"`
	Type          string `json:"type"`
	Role          string `json:"role"`
}

type GatewayResponse struct {
	DeezerError []DeezerError `json:"error"`
	Results     struct {
		User struct {
			Options struct {
				LicenseToken string `json:"license_token"`
			}
		}
	}
}

var client Client

type Client struct {
	http         *http.Client
	apiBaseURL   string
	webBaseURL   string
	mediaBaseURL string
	arl          string
}

func InitClient(arl string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		http: &http.Client{
			Jar: jar,
		},
		apiBaseURL:   "https://api.deezer.com",
		webBaseURL:   "https://www.deezer.com",
		mediaBaseURL: "https://media.deezer.com",
		arl:          arl,
	}
}

func joinURL(baseURL, path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")
}

var softErrors = []int{
		http.StatusInternalServerError,
		http.StatusNotImplemented,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusHTTPVersionNotSupported,
		http.StatusVariantAlsoNegotiates,
		http.StatusInsufficientStorage,
		http.StatusLoopDetected,
		http.StatusNotExtended,
		http.StatusNetworkAuthenticationRequired,
	}


func (c *Client) getJSON(ctx context.Context, baseURL, path string, v any) error {

	u := joinURL(baseURL, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}


	if c.arl != "" {
		req.AddCookie(&http.Cookie{Name: "arl", Value: c.arl})
	}

	response, err := c.http.Do(req)
	if err != nil {
		if response != nil {
			response.Body.Close()
		}
		return fmt.Errorf("GET %s: %w", u, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		switch {
		case response.StatusCode == http.StatusUnauthorized ||
			response.StatusCode == http.StatusForbidden:
			return fatal(fmt.Errorf("GET %s: %s", u, response.Status))

		case response.StatusCode == http.StatusNotFound:
			return fmt.Errorf("GET %s: %s", u, response.Status)

		case slices.Contains(softErrors, response.StatusCode):
			return fmt.Errorf("GET %s: %s", u, response.Status)

		default:
			return fmt.Errorf("GET %s: %s", u, response.Status)
		}
	}

	if err := json.NewDecoder(response.Body).Decode(v); err != nil {
		return fmt.Errorf("decode %s: %w", u, err)
	}

	return nil
}

func (c *Client) GetTrack(ctx context.Context, trackID int) (Track, error) {
	path := fmt.Sprintf("/track/%d", trackID)
	var track Track
	err := c.getJSON(ctx, c.apiBaseURL, path, &track)
	if err != nil {
		return Track{}, err
	}

	return track, nil
}

func (c *Client) GetAlbum(ctx context.Context, id int) (Album, error) {
	var album Album
	path := fmt.Sprintf("/album/%d", id)
	if err := c.getJSON(ctx, c.apiBaseURL, path, &album); err != nil {
		return Album{}, err
	}

	return album, nil
}

func (c *Client) GetUser(ctx context.Context, uid int) (User, error) {
	path := fmt.Sprintf("/user/%d", uid)
	var user User
	if err := c.getJSON(ctx, c.apiBaseURL, path, &user); err != nil {
		return User{}, err
	}

	return user, nil
}

func (c *Client) GetPlaylist(ctx context.Context, id int) (Playlist, error) {
	var playlist Playlist
	path := fmt.Sprintf("/playlist/%d", id)
	if err := c.getJSON(ctx, c.apiBaseURL, path, &playlist); err != nil {
		return Playlist{}, err
	}

	return playlist, nil
}

func (c *Client) fetchPlaylist(ctx context.Context, id int) (FetchedDataPlaylist, error) {
	playlist, err := c.GetPlaylist(ctx, id)
	if err != nil {
		return FetchedDataPlaylist{}, err
	}

	var tracks []string

	for _, track := range playlist.Tracks.Data {
		tracks = append(tracks, fmt.Sprintf("%s - %s", track.Artist.Name, track.Title))
	}

	var data FetchedDataPlaylist
	data.PlaylistName = playlist.Title
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

type Page[T any] struct {
	Data  []T    `json:"data"`
	Total int    `json:"total"`
	Next  string `json:"next"`
}

func (c *Client) fetchAlbums(ctx context.Context, uid int) ([]Album, error) {
	var albums []Album
	next := fmt.Sprintf("/user/%d/albums", uid)

	for next != "" {
		var page Page[Album]
		if err := c.getJSON(ctx, c.apiBaseURL, next, &page); err != nil {
			return nil, fmt.Errorf("get user %d albums: %w", uid, err)
		}
		albums = append(albums, page.Data...)
		next = page.Next
	}
	return albums, nil
}

func (c *Client) GetUserTracks(ctx context.Context, uid int) ([]Track, error) {
	var tracks []Track
	next := fmt.Sprintf("/user/%d/tracks", uid)

	for next != "" {
		var page Page[Track]
		if err := c.getJSON(ctx, c.apiBaseURL, next, &page); err != nil {
			return []Track{}, err
		}
		tracks = append(tracks, page.Data...)

		next = page.Next
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

func (c *Client) getLicenseToken(ctx context.Context, arl string) (string, error) {
	path := "/ajax/gw-light.php?method=deezer.getUserData&input=3&api_version=1.0&api_token="
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.webBaseURL+path, nil)

	if err != nil {
		return "", err
	}

	req.Header.Set("Cookie", "arl="+arl)
	resp, err := c.http.Do(req)
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

func (c *Client) getTrackToken(arl string, trackID int) (string, error) {
	apiEndpoint := fmt.Sprintf("https://api.deezer.com/track/%d", trackID)
	req, err := http.NewRequest("GET", apiEndpoint, http.NoBody)
	if err != nil {
		return "", nil
	}

	req.Header.Set("Cookie", "arl="+arl+"; comeback=1")
	resp, err := c.http.Do(req)
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
		LicenseToken string   `json:"license_token"`
		Media        []media  `json:"media"`
		TrackTokens  []string `json:"track_tokens"`
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

func (c *Client) getTrackURL(licenseToken, trackToken string) (string, error) {
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
		Errors []DeezerError `json:"errors"`
	}

	firstMediaURL := func(resp getURLResponse) (string, error) {
		for _, dataItem := range resp.Data {
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
		if len(resp.Errors) > 0 {
			e:= resp.Errors[0]
			return "", fmt.Errorf("get_url: %s (code %d)", e.Err.Message, e.Err.Code)
		}

		return "", errors.New("get_url: no media sources")
	}
	

	payload, err := createMediaPayload(licenseToken, trackToken)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", joinURL(c.mediaBaseURL, "/v1/get_url"), bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return "", fmt.Errorf("get_url: %s: %s", resp.Status, body)
	}

	var result getURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return firstMediaURL(result)

}

var unsafeFilenameChars = strings.NewReplacer(
	"/", "_",
	"\\", "_",
	":", "_",
	"*", "_",
	"?", "_",
	"\"", "_",
	"<", "_",
	">", "_",
	"|", "_",
)

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	name = unsafeFilenameChars.Replace(name)
	name = strings.Trim(name, ". ")

	if name == "" {
		return "track"
	}

	runes := []rune(name)
	if len(runes) > 200 {
		name = string(runes[:200])
	}

	return name
}

func getArtists(track Track) string {
	if len(track.Contributors) == 0 {
		return track.Artist.Name
	}
	var artistsBuilder strings.Builder

	for i, contributor := range track.Contributors {
		artistsBuilder.WriteString(contributor.Name)
		if i < len(track.Contributors) - 1 {
			artistsBuilder.WriteString(", ")
		}
	}
	artists := artistsBuilder.String()
	return artists
}
func formatTrackName(track Track) string {
	artists := getArtists(track)
	filename := fmt.Sprintf("%s - %s", artists, track.Title)
	filename = strings.TrimSpace(filename)
	filename = unsafeFilenameChars.Replace(filename)
	filename = strings.Trim(filename, ". ")
	if filename == "" {
		return "untitled"
	}

	runes := []rune(filename)
	if len(runes) > 200 {
		filename = string(runes[:200])
	}
	return filename
}

func (d *Downloader) fetchAndDecrypt(
	ctx context.Context,
	trackURL string,
	trackInfo Track,
	filePath io.Writer) error {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, trackURL, nil)
	if err != nil {
		return err
	}
	resp, err := d.client.http.Do(req)
	if err != nil {
		return fmt.Errorf("GET media: %w", err)
	}

	defer resp.Body.Close()

	trackID := fmt.Sprintf("%d", trackInfo.ID)
	key := calcBlowfishKey(trackID)

	cw := &countingWriter{w: filePath}

	if err := decryptFile(resp.Body, key, cw); err != nil {
		return fmt.Errorf("Failed to decrypt: %w", err)
	}

	if resp.ContentLength >= 0 && cw.n != resp.ContentLength {
		return fmt.Errorf("Truncated media: got %d of %d bytes", cw.n, resp.ContentLength)
	}

	return nil
}

type countingWriter struct {
	w io.Writer
	n int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	return n, err
}

func (d *Downloader) GetCover(ctx context.Context, albumID int) ([]byte, error) {
	album, err := d.getAlbum(ctx, albumID)
	if err != nil {
		return nil, err
	}

	var coverURL string
	if album.CoverBig != "" {
		coverURL = album.CoverBig
	} else if album.CoverMedium != "" {
		coverURL = album.CoverMedium
	} else if album.CoverSmall != "" {
		coverURL = album.CoverSmall
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, coverURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cover GET: %s", resp.Status)
	}

	return io.ReadAll(io.LimitReader(resp.Body, 10<<20))

}


func (d *Downloader) addMetadata(ctx context.Context, path string, track Track) error {

	artists := make([]string, 0, len(track.Contributors))
	for _, contributor := range track.Contributors {
		artists = append(artists, contributor.Name)
	}

	album, err := d.getAlbum(ctx, track.Album.ID)

	genre := ""
	if len(album.Genres.Data) > 0 {
		genre = album.Genres.Data[0].Name
	}

	releaseDate, err := parseReleaseDate(track)

	tags := map[string][]string{
		taglib.Artist:      artists,
		taglib.Title:       {track.Title},
		taglib.Album:       {track.Album.Title},
		taglib.AlbumArtist: {track.Album.Artist.Name},
		taglib.TrackNumber: {fmt.Sprintf("%d", track.TrackPosition)},
		taglib.ISRC:        {track.ISRC},
		taglib.BPM:         {fmt.Sprintf("%f", track.BPM)},
		taglib.Genre:       {genre},
		taglib.Label:       {album.Label},
		"Year":             {fmt.Sprintf("%d", releaseDate.Year())},
		"Gain":             {fmt.Sprintf("%f", track.Gain)},
	}

	cover, err := d.GetCover(ctx, album.ID)
	if err != nil {
		return fmt.Errorf("Get cover: %w", err)
	}

	if err = taglib.WriteTags(path, tags, 0); err != nil {
		return err
	}

	err = taglib.WriteImageOptions(
		path,
		cover, 0,
		"Front Cover",
		"Front artwork",
		"image/jpeg")

	if err != nil {
		return err
	}

	return nil
}

func (d *Downloader) getAlbum(ctx context.Context, id int) (Album, error) {
	d.mu.Lock()
	if album, ok := d.albumCache[id]; ok {
		d.mu.Unlock()
		return album, nil
	}
	d.mu.Unlock()

	album, err := d.client.GetAlbum(ctx, id)
	if err != nil {
		return Album{}, nil
	}

	d.mu.Lock()
	d.albumCache[id] = album
	d.mu.Unlock()

	return album, nil
}

type Downloader struct {
	client       *Client
	outDir       string
	licenseToken string
	force        bool
	mu           sync.Mutex
	albumCache   map[int]Album
}

func (d *Downloader) DownloadTrack(
	ctx context.Context,
	trackID int,
) (string, error) {
	return d.downloadTrack(ctx, trackID, formatTrackName)
}

func (d *Downloader) downloadTrack(ctx context.Context,
	trackID int,
	filename FilenameFunc,
) (string, error) {
	if trackID <= 0 {
		return "", fmt.Errorf("Invalid track ID: %d", trackID)
	}

	track, err := d.client.GetTrack(ctx, trackID)
	if err != nil {
		return "", fmt.Errorf("Get track %d: %w", trackID, err)
	}

	finalPath := filepath.Join(
		d.outDir,
		filename(track)+".mp3",
	)

	if _, err := os.Stat(finalPath); err == nil {
		fmt.Printf("Skipping: '%s' - file exists\n", finalPath)
		return "", nil
	}
	fmt.Printf("Downloading: '%s - %s'...\n", track.Artist.Name, track.Title)

	trackToken := track.TrackToken

	trackURL, err := d.client.getTrackURL(d.licenseToken, trackToken)
	if err != nil {
		return "", fmt.Errorf("Failed to get URL for track with id %d: %w", trackID, err)
	}

	tmp, err := os.CreateTemp(d.outDir, ".part-*")
	if err != nil {
		return "", fatal(fmt.Errorf("Failed to create temp file: %w", err))
	}

	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	if err := d.fetchAndDecrypt(ctx, trackURL, track, tmp); err != nil {
		return "", fmt.Errorf("Failed to save track with id %d: %w", trackID, err)
	}

	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("Close track %d: %w", trackID, err)
	}

	if err := d.addMetadata(ctx, tmpPath, track); err != nil {
		return "", fmt.Errorf("Write metadata: %w", err)
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		return "", fatal(fmt.Errorf("Rename track %d: %w", trackID, err))
	}

	return finalPath, nil
}

func (d *Downloader) downloadAll(ctx context.Context,
	tracks []Track,
	filename FilenameFunc,
) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(3)

	var (
		mu       sync.Mutex
		softErrs []error
	)

	for _, track := range tracks {
		if ctx.Err() != nil {
			break
		}

		g.Go(func() error {
			if _, err := d.downloadTrack(ctx, track.ID, filename); err != nil {
				if isFatal(err) || ctx.Err() != nil {
					return fmt.Errorf("track '%s - %s' (ID %d): %w", track.Artist.Name, track.Title, track.ID, err)
				}

				log.Printf("track '%s - %s' (ID %d): %v", track.Artist.Name, track.Title, track.ID, err)

				mu.Lock()

				softErrs = append(softErrs, fmt.Errorf("track '%s - %s': %w", track.Artist.Name, track.Title, err))
				mu.Unlock()
			}
			return nil
		})
	}
	fatalErr := g.Wait()
	softErr := errors.Join(softErrs...)
	if fatalErr != nil {
		return errors.Join(fatalErr, softErr)
	}
	return softErr
}

func (d *Downloader) DownloadPlaylist(ctx context.Context, id int) error {
	playlist, err := d.client.GetPlaylist(ctx, id)
	if err != nil {
		return err
	}
	if err := d.downloadAll(ctx, playlist.Tracks.Data, formatTrackName); err != nil {
		return err
	}
	return nil
}

func (d *Downloader) DownloadFavorite(ctx context.Context, uid int) error {
	tracks, err := d.client.GetUserTracks(ctx, uid)
	if err != nil {
		return err
	}
	if err := d.downloadAll(ctx, tracks, formatTrackName); err != nil {
		return err
	}
	return nil
}

type FilenameFunc func(Track) string

func hasMultipleDiscs(tracks []Track) bool {
	for _, track := range tracks {
		if track.DiskNumber > 1 {
			return true
		}
}
	return false
}

func albumTrackLabel(track Track, multiDisc bool) string {
	discs := map[int]string{
		1: "A",
		2: "B",
		3: "C",
		4: "D",
	}
		if multiDisc {
			disc := track.DiskNumber
			if disc <= 0 {
				disc = 1
			}
			return fmt.Sprintf("%s%02d.", discs[disc], track.TrackPosition)
		}
		return fmt.Sprintf("%02d.", track.TrackPosition)
}


func albumFilenameFunc(tracks []Track) FilenameFunc {
	multiDisc := hasMultipleDiscs(tracks)

	return func(track Track) string {
		name := formatTrackName(track)
		trackLabel := albumTrackLabel(track, multiDisc)
		if trackLabel == "" {
			return name
		}
		return trackLabel + " " + name
	}
}

func (d *Downloader) DownloadAlbum(ctx context.Context, id int) error {
	var tracks []Track

	next := fmt.Sprintf("/album/%d/tracks", id)

	for next != "" {
		var page Page[Track]

	if err := d.client.getJSON(ctx, d.client.apiBaseURL, next, &page); err != nil {
		return err
	}

	tracks = append(tracks, page.Data...)
	next = page.Next
	}

	return d.downloadAll(ctx, tracks, albumFilenameFunc(tracks))
}

func InitDownloader(ctx context.Context, c *Client, outDir string, force bool) (*Downloader, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil && !errors.Is(err, fs.ErrExist) {
		return nil, fmt.Errorf("Failed to create directory: %w", err)
	}

	licenseToken, err := c.getLicenseToken(ctx, c.arl)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve license token: %w", err)
	}

	if licenseToken == "" {
		return nil, fatal(errors.New("Empty license token; check ARL"))
	}

	return &Downloader{
		client:       c,
		outDir:       outDir,
		licenseToken: licenseToken,
		force:        force,
		albumCache:   make(map[int]Album),
	}, nil

}

func printUsage() {
	name := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, `Usage:
	  %[1]s search track <query>
	  %[1]s search album <query>
	  %[1]s search playlist <query>
	  %[1]s download track [flags] <track-id>
	  %[1]s download album [flags] <album-id>
	  %[1]s download playlist [flags] <playlist-id>
	  %[1]s download favorites [flags] <user-id>
	  %[1]s show track <track-id>
	  %[1]s show album <album-id>
	  %[1]s show playlist <playlist-id>
	  %[1]s help

	Flags for download commands:
	  -out string    output directory
	  -force         overwrite existing files
	  `, name)
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		printUsage()
		return errors.New("missing command")
	}

	switch args[0] {
	case "download":
		return runDownload(ctx, args[1:])
	case "search":
		return runSearch(ctx, args[1:])
	case "show":
		return runShow(ctx, args[1:])
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("Unknown command %q", args[0])
	}
}

func runDownload(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("download: expected track, album, playlist or favorites")
	}
	
	switch args[0] {
	case "track":
		return runDownloadTrack(ctx, args[1:])
	case "album":
		return runDownloadAlbum(ctx, args[1:])
	case "playlist":
		return runDownloadPlaylist(ctx, args[1:])
	case "favorites":
		return runDownloadFavorites(ctx, args[1:])
	default:
		return fmt.Errorf("unknown download target %q", args[0])
	}
}

func parseSingleID(fs *flag.FlagSet, args []string, kind string) (int, error) {
	if err := fs.Parse(args); err != nil {
		return 0, err
	}
	
	if fs.NArg() != 1 {
		fs.Usage()
		return 0, fmt.Errorf("%s: expected exactly one ID", kind)
	}

	id, err := strconv.Atoi(fs.Arg(0))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid %s ID %q", kind, fs.Arg(0))
	}

	return id, nil
}

func defaultDir(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func addDownloadFlags(fs *flag.FlagSet, opts *downloadOptions) {
	fs.StringVar(&opts.outDir, "out", "", "output directory")
	fs.BoolVar(&opts.force, "force", false, "overwrite existing files")
}

func runDownloadTrack(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("download track", flag.ContinueOnError)
	var opts downloadOptions
	addDownloadFlags(fs, &opts)

	id, err := parseSingleID(fs, args, "track")
	if err != nil {
		return err
	}

	client := InitClient(os.Getenv("DEEZER_ARL"))
	downloader, err := InitDownloader(ctx, client, defaultDir(opts.outDir, "downloads"), opts.force)

	if err != nil {
		return err
	}
	path, err := downloader.DownloadTrack(ctx, id)
	if err != nil {
		return err
	}
	fmt.Printf("Saved: %s", path)

	return nil
}

func runDownloadAlbum(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("download album", flag.ContinueOnError)
	var opts downloadOptions
	addDownloadFlags(fs, &opts)

	id, err := parseSingleID(fs, args, "album")
	if err != nil {
		return err
	}

	client := InitClient(os.Getenv("DEEZER_ARL"))
	album, err := client.GetAlbum(ctx, id)

	if err != nil {
		return err
	}
	dir := opts.outDir
	if dir == "" {
		dir = filepath.Join("downloads", "albums",
		sanitizeFilename(album.Artist.Name + " - " + album.Title))
	}
	downloader, err := InitDownloader(ctx, client, dir, opts.force)
	if err != nil {
		return err
	}

	if err := downloader.DownloadAlbum(ctx, id); err != nil {
		return err
	}

	fmt.Printf("Album saved to %s\n", dir)
	return nil
}

func runDownloadPlaylist(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("download playlist", flag.ContinueOnError)
	var opts downloadOptions
	addDownloadFlags(fs, &opts)

	id, err := parseSingleID(fs, args, "playlist")
	if err != nil {
		return err
	}

	client := InitClient(os.Getenv("DEEZER_ARL"))
	playlist, err := client.GetPlaylist(ctx, id)

	if err != nil {
		return err
	}
	dir := opts.outDir
	if dir == "" {
		dir = filepath.Join("downloads", "playlists",
		sanitizeFilename(playlist.Title))
	}
	downloader, err := InitDownloader(ctx, client, dir, opts.force)
	if err != nil {
		return err
	}

	if err := downloader.DownloadPlaylist(ctx, id); err != nil {
		return err
	}

	fmt.Printf("Playlist saved to %s\n", dir)
	return nil
}

func runDownloadFavorites(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("download favorites", flag.ContinueOnError)
	var opts downloadOptions
	addDownloadFlags(fs, &opts)

	id, err := parseSingleID(fs, args, "favorites")
	if err != nil {
		return err
	}

	client := InitClient(os.Getenv("DEEZER_ARL"))

	dir := opts.outDir
	if dir == "" {
		dir = filepath.Join("downloads", "favorites")
	}

	downloader, err := InitDownloader(ctx, client, dir, opts.force)
	if err != nil {
		return err
	}

	if err := downloader.DownloadFavorite(ctx, id); err != nil {
		return err
	}

	fmt.Printf("Saved to %s\n", dir)
	return nil
}


type downloadOptions struct {
	outDir	string
	force bool
}

func search[T any](
	ctx context.Context,
	c *Client,
	kind, query string,
	limit int,
) ([]T, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("empty search query")
	}
	if limit <= 0 {
		limit = 25
	}
	q := url.Values{}
	q.Set("q", query)
	q.Set("limit", strconv.Itoa(limit))

	path := "/search/" + kind + "?" + q.Encode()

	var page Page[T]
	if err := c.getJSON(ctx, c.apiBaseURL, path, &page); err != nil {
		return nil, fmt.Errorf("search %s %q: %w", kind, query, err)
	}

	return page.Data, nil
}

func (c *Client) SearchTrack(ctx context.Context, query string, limit int) ([]Track, error) {
	return search[Track](ctx, c, "track", query, limit)
}

func (c *Client) SearchAlbum(ctx context.Context, query string, limit int) ([]Album, error) {
	return search[Album](ctx, c, "album", query, limit)
}

func (c *Client) SearchPlaylist(ctx context.Context, query string, limit int) ([]Playlist, error) {
	return search[Playlist](ctx, c, "playlist", query, limit)
}

func (c *Client) SearchArtist(ctx context.Context, query string, limit int) ([]Artist, error) {
	return search[Artist](ctx, c, "artist", query, limit)
}


func runSearch(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("search: expected track, album, playlist, or artist")
	}

	switch args[0] {
	case "track":
		return runSearchTrack(ctx, args[1:])
	case "album":
		return runSearchAlbum(ctx, args[1:])
	case "playlist":
		return runSearchPlaylist(ctx, args[1:])
	default:
		return fmt.Errorf("unknown search type %q", args[0])
	}
}

func parseSearchQuery(fs *flag.FlagSet, args []string, kind string) (string, error) {
	if err := fs.Parse(args); err != nil {
		return "", err
	}
	if fs.NArg() == 0 {
		fs.Usage()
		return "", fmt.Errorf("search %s: query required", kind)
	}
	return strings.Join(fs.Args(), " "), nil
}

type Media interface {
	HasExplicitLyrics() bool
	GetReleaseDate() string
	GetDuration()	int
}

func (t Track) HasExplicitLyrics() bool {
	return t.ExplicitLyrics
}

func (a Album) HasExplicitLyrics() bool {
	return a.ExplicitLyrics
}

func (p Playlist) HasExplicitLyrics() bool {
	return p.ExplicitLyrics
}

func (t Track) GetReleaseDate() string {
	return t.ReleaseDate
}


func (a Album) GetReleaseDate() string {
	return a.ReleaseDate
}

func (p Playlist) GetReleaseDate() string {
	return p.ReleaseDate
}

func (t Track) GetDuration() int {
	return t.Duration
}

func (a Album) GetDuration() int {
	return a.Duration
}

func (p Playlist) GetDuration() int {
	return p.Duration
}

func explicitMarker(media Media) string {
	if media.HasExplicitLyrics() {
		return " [E]"
	}
	return ""
}

func parseReleaseDate(media Media) (time.Time, error) {
	const dateFormat = "2006-01-02"
	releaseDate, err := time.Parse(dateFormat, media.GetReleaseDate())
	if err != nil {
		return time.Time{}, err
	}
	return releaseDate, nil
}

func formatDuration(media Media) string {
	d := time.Duration(media.GetDuration()) * time.Second
	return fmt.Sprintf("%02d:%02d", int(d.Minutes()), int(d.Seconds())%60)

}

func formatTrackSearchResult(track Track) string {
	artists := getArtists(track)
	duration := formatDuration(track)
	return fmt.Sprintf(
		"%d\t%s - %s (%s)%s",
		track.ID,
		artists,
		track.Title,
		duration,
		explicitMarker(track),
	)
}

func formatAlbumSearchResult(album Album, releaseDate string) string {
	var date string

	if releaseDate != "" {
		date = fmt.Sprintf(" (%s)", releaseDate)
	}

	return fmt.Sprintf(
		"%d\t%s - %s%s%s",
		album.ID,
		album.Artist.Name,
		album.Title,
		date,
		explicitMarker(album),
	)
}

func formatPlaylistSearchResult(playlist Playlist, creationDate string) string {
	var date string
	if creationDate != "" {
		date = fmt.Sprintf("%s", creationDate)
	}

	return fmt.Sprintf(
		"%d\t%s (created by: %s, %s) (tracks: %d)",
		playlist.ID,
		playlist.Title,
		playlist.User.Name,
		date,
		playlist.NbTracks,
	)
}

func runSearchTrack(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("search track", flag.ContinueOnError)
	limit := fs.Int("limit", 25, "maximum number of results")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: dzscrape search track [flags] <query>")
		fs.PrintDefaults()
	}

	query, err := parseSearchQuery(fs, args, "track")
	if err != nil {
		return err
	}

	client := InitClient(os.Getenv("DEEZER_ARL"))

	tracks, err := client.SearchTrack(ctx, query, *limit)
	if err != nil {
		return err
	}

	for _, track := range tracks {
		fmt.Println(formatTrackSearchResult(track))
	}
	return nil
}

func runSearchAlbum(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("search album", flag.ContinueOnError)
	limit := fs.Int("limit", 25, "maximum number of results")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: dzscrape search album [flags] <query>")
		fs.PrintDefaults()
	}

	query, err := parseSearchQuery(fs, args, "album")
	if err != nil {
		return err
	}

	client := InitClient(os.Getenv("DEEZER_ARL"))

	albums, err := client.SearchAlbum(ctx, query, *limit)
	if err != nil {
		return err
	}

	dates := make(map[int]string)
	for _, album := range albums {
		date := dates[album.ID]
		if date == ""{
			album, err := client.GetAlbum(ctx, album.ID)
			if err == nil {
				date = album.ReleaseDate
				dates[album.ID] = date
			}
		}

		fmt.Println(formatAlbumSearchResult(album, date))
	}
	return nil
}

func runSearchPlaylist(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("search playlist", flag.ContinueOnError)
	limit := fs.Int("limit", 25, "maximum number of results")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: dzscrape search playlist [flags] <query>")
		fs.PrintDefaults()
	}

	query, err := parseSearchQuery(fs, args, "playlist")
	if err != nil {
		return err
	}

	client := InitClient(os.Getenv("DEEZER_ARL"))

	playlists, err := client.SearchPlaylist(ctx, query, *limit)
	if err != nil {
		return err
	}

	for _, playlist := range playlists {

		fmt.Println(formatPlaylistSearchResult(playlist, playlist.CreationDate))
	}

	return nil
}

func formatTrackShowResult(track Track) string {
	genre := "unknown"
	if len(track.Album.Genres.Data) > 0 {
		genre = track.Album.Genres.Data[0].Name
	}

	return fmt.Sprintf(
`Title: %s
Artist: %s
Contributors: %s
Album: %s
Duration: %s
Release date: %s
Genre: %s
BPM: %f
Rank: %d
Explicit lyrics: %t
Label: %s
Cover (small): %s
Cover (medium): %s
Cover (big): %s
Cover (XL): %s
`,
	track.Title,
	track.Artist.Name,
	getArtists(track),
	track.Album.Title,
	formatDuration(track),
	track.ReleaseDate,
	genre,
	track.BPM,
	track.Rank,
	track.ExplicitLyrics,
	track.Album.Label,
	track.Album.CoverSmall,
	track.Album.CoverMedium,
	track.Album.CoverBig,
	track.Album.CoverXL,
	)
}

func formatAlbumTracklist(tracks []Track) string {
	multiDisc := hasMultipleDiscs(tracks)

	var b strings.Builder

	for _, track := range tracks {
		if label := albumTrackLabel(track, multiDisc); label != "" {
			b.WriteString(label)
			b.WriteByte(' ')
		}

		b.WriteString(getArtists(track))
		b.WriteString(" - ")
		b.WriteString(track.Title)

		if track.Duration > 0 {
			b.WriteString(" (")
			b.WriteString(formatDuration(track))
			b.WriteByte(')')
		}

		b.WriteString(explicitMarker(track))
		b.WriteByte('\n')
	}

	return strings.TrimRight(b.String(), "\n")
}
func formatAlbumShowResult(client *Client, ctx context.Context, album Album) string {
	genre := "unknown"
	if len(album.Genres.Data) > 0 {
		genre = album.Genres.Data[0].Name
	}

	var tracks []Track
	var tracklist string
	next := fmt.Sprintf("/album/%d/tracks", album.ID)
	for next != "" {
		var page Page[Track]

	if err := client.getJSON(ctx, client.apiBaseURL, next, &page); err != nil {
		tracklist = "could not fetch"
	}

	tracks = append(tracks, page.Data...)
	next = page.Next
	}
	tracklist = formatAlbumTracklist(tracks)
	return fmt.Sprintf(
`Title: %s
Artist: %s
Duration: %s
Release date: %s
Genre: %s
Tracks: %d
Explicit lyrics: %t
Label: %s
Tracklist:
%s
Cover (small): %s
Cover (medium): %s
Cover (big): %s
Cover (XL): %s
`,
	album.Title,
	album.Artist.Name,
	formatDuration(album),
	album.ReleaseDate,
	genre,
	album.NbTracks,
	album.ExplicitLyrics,
	album.Label,
	tracklist,
	album.CoverSmall,
	album.CoverMedium,
	album.CoverBig,
	album.CoverXL,
	)
}

func formatPlaylistTracklist(playlist Playlist) string {
	var tracklist string
	for _, track := range playlist.Tracks.Data {
		artists := getArtists(track)
		tracklist += fmt.Sprintf("%s - %s\n", artists, track.Title)
	}
	return tracklist
}

func formatPlaylistShowResult(playlist Playlist) string {
	return fmt.Sprintf(
`Title: %s
Creator: %s
Duration: %s
Creation date: %s
Tracklist:
%s
`,
	playlist.Title,
	playlist.Creator.Name,
	formatDuration(playlist),
	playlist.CreationDate,
	formatPlaylistTracklist(playlist),
	)
}

func runShow(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("show: expected track, album, playlist, or artist")
	}

	switch args[0] {
	case "track":
		return runShowTrack(ctx, args[1:])
	case "album":
		return runShowAlbum(ctx, args[1:])
	case "playlist":
		return runShowPlaylist(ctx, args[1:])
	default:
		return fmt.Errorf("unknown show option %q", args[0])
	}
}

func runShowTrack(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("show track", flag.ContinueOnError)
	id, err := parseSingleID(fs, args, "track")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: dzscrape show track <track-id>")
	}
	client := InitClient(os.Getenv("DEEZER_ARL"))
	track, err := client.GetTrack(ctx, id)
	if err != nil {
		return err
	}
	fmt.Println(formatTrackShowResult(track))
	return nil
}

func runShowAlbum(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("show album", flag.ContinueOnError)
	id, err := parseSingleID(fs, args, "album")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: dzscrape show album <album-id>")
	}
	client := InitClient(os.Getenv("DEEZER_ARL"))
	album, err := client.GetAlbum(ctx, id)
	if err != nil {
		return err
	}
	fmt.Println(formatAlbumShowResult(client, ctx, album))
	return nil
}

func runShowPlaylist(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("show playlist", flag.ContinueOnError)
	id, err := parseSingleID(fs, args, "playlist")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: dzscrape show playlist <playlist-id>")
	}
	client := InitClient(os.Getenv("DEEZER_ARL"))
	playlist, err := client.GetPlaylist(ctx, id)
	if err != nil {
		return err
	}
	fmt.Println(formatPlaylistShowResult(playlist))
	return nil
}

type fatalError struct {
	err error
}

func (e *fatalError) Error() string { return e.err.Error() }
func (e *fatalError) Unwrap() error { return e.err }

func fatal(err error) error {
	if err == nil {
		return nil
	}
	return &fatalError{err: err}
}

func isFatal(err error) bool {
	var target *fatalError
	return errors.As(err, &target)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
		os.Exit(1)
	}
}
