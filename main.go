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
	_"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
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
	Creator       Creator `json:"creator"`
	Tracks	struct{
		Data []Track	`json:"data"`
	}	`json:"tracks"`
}

type Creator struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Tracklist string `json:"tracklist"`
	Type      string `json:"type"`
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
	ID          int		   `json:"id"`
	Title       string	   `json:"title"`
	Link        string	   `json:"link"`
	Upc         string	   `json:"upc"`
	Cover       string	   `json:"cover"`
	CoverSmall  string	   `json:"cover_small"`
	CoverMedium string	   `json:"cover_medium"`
	CoverBig    string	   `json:"cover_big"`
	CoverXL     string	   `json:"cover_xl"`
	MD5Image    string	   `json:"md5_image"`
	Genres      struct {
		Data []GenreData `json:"data"`
	} `json:"genres"`
	Label     string 	`json:"label"`
	Tracklist string 	`json:"tracklist"`
	Type      string 	`json:"type"`
	Tracks	  struct {
		Data	[]Track	`json:"data"`
	}
	Artist    Artist 	`json:"artist"`
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
	ID                    int            `json:"id"`
	Readable              bool           `json:"readable"`
	Title                 string         `json:"title"`
	TitleShort            string         `json:"title_short"`
	TitleVersion          string         `json:"title_version"`
	ISRC                  string         `json:"isrc"`
	Link                  string         `json:"link"`
	Share                 string         `json:"share"`
	Duration              int            `json:"duration"`
	TrackPosition         int            `json:"track_position"`
	DiskNumber            int            `json:"disk_number"`
	Rank                  int            `json:"rank"`
	ReleaseDate           string         `json:"release_date"`
	ExplicitLyrics        bool           `json:"explicit_lyrics"`
	ExplicitContentLyrics int            `json:"explicit_content_lyrics"`
	ExplicitContentCover  int            `json:"explicit_content_cover"`
	Preview               string         `json:"preview"`
	BPM                   float64        `json:"bpm"`
	Gain                  float64        `json:"gain"`
	AvailableCountries    []string       `json:"available_countries"`
	Contributors          []Contributor `json:"contributors"`
	MD5Image              string         `json:"md5_image"`
	TrackToken            string         `json:"track_token"`
	Artist                Artist         `json:"artist"`
	Album                 Album          `json:"album"`
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
		return fmt.Errorf("get %s: %w", u, err)
	} else if response.StatusCode != http.StatusOK {
		defer response.Body.Close()
		return fmt.Errorf("get %s: unexpected status %s", u, response.Status)
	}

	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(v)
	if err != nil {
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

func (c *Client) GetAlbum(ctx context.Context, id int) (Album, error){
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

func (c *Client) GetPlaylist(ctx context.Context, id int) (Playlist, error){
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
	Data	[]T		`json:"data"`
	Total	int		`json:"total"`
	Next	string	`json:"next"`
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.webBaseURL + path, nil)

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
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}
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

	for i, contributor := range trackInfo.Contributors {
		artistsBuilder.WriteString(contributor.Name)
		if i < len(trackInfo.Contributors)-1 {
			artistsBuilder.WriteString(", ")
		}
	}

	artists := artistsBuilder.String()

	filename := fmt.Sprintf("%s - %s", artists, trackInfo.Title)
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

	if err := decryptFile(resp.Body, key, cw); err != nil{
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

	const dateFormat = "2006-01-02"
	releaseDate, err := time.Parse(dateFormat, track.ReleaseDate)
	if err != nil {
		return err
	}

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
	client			*Client
	outDir			string
	licenseToken	string
	force			bool
	mu				sync.Mutex
	albumCache		map[int]Album
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
		filename(track) + ".mp3",
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
	if err != nil{
		return "", fmt.Errorf("Failed to create temp file: %w", err)
	}

	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	if err := d.fetchAndDecrypt(ctx, trackURL, track, tmp); err !=nil {
		return "", fmt.Errorf("Failed to save track with id %d: %w", trackID, err)
	}

	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("Close track %d: %w", trackID, err)
	}

	if err := d.addMetadata(ctx, tmpPath, track); err != nil {
		return "", fmt.Errorf("Write metadata: %w", err)
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
			return "", fmt.Errorf("Rename track %d: %w", trackID, err)
		}


	return finalPath, nil
}

func (d *Downloader) downloadAll(ctx context.Context,
	tracks []Track,
	filename FilenameFunc,
)	error {
	var (
		g	errgroup.Group
		mu sync.Mutex
		errs []error
	)

	g.SetLimit(3)
	
	for _, track := range tracks {
		if ctx.Err() != nil {
			break
		}
		g.Go(func() error {
			if _, err := d.downloadTrack(ctx, track.ID, filename); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("track '%s - %s': %w", track.Artist.Name, track.Title, err))
				mu.Unlock()
			}
			return nil
		})
	}

	_ = g.Wait()
	return errors.Join(errs...)
}

func (d *Downloader) DownloadPlaylist(ctx context.Context, id int) error {
	playlist, err := d.client.GetPlaylist(ctx, id); if err != nil {
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

func albumFilenameFunc(tracks []Track) FilenameFunc {
	return func(track Track) string {
		name := formatTrackName(track)
		if track.TrackPosition <= 0 {
			return name
		}
		return fmt.Sprintf("%02d. %s", track.TrackPosition, name)
	}
}

func (d *Downloader) DownloadAlbum(ctx context.Context, id int) error {
	album, err := d.getAlbum(ctx, id)
	if err != nil {
		return err
	}
	return d.downloadAll(ctx, album.Tracks.Data, albumFilenameFunc(album.Tracks.Data))
}


func InitDownloader(ctx context.Context, c *Client, outDir string, force bool) (*Downloader, error) {
	if err := os.Mkdir(outDir, 0o755); err != nil && !errors.Is(err, fs.ErrExist) {
		return nil, fmt.Errorf("Failed to create directory: %w", err)
	}

	licenseToken, err := c.getLicenseToken(ctx, c.arl)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve license token: %w", err)
	}

	if licenseToken == "" {
			return nil, errors.New("empty license token; check ARL")
		}

	return &Downloader{
		client: c,
		outDir: outDir,
		licenseToken: licenseToken,
		force:	force,
		albumCache:	make(map[int]Album),
	}, nil

}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("No commands were provided")
	}

	switch args[0] {
	case "track":
		return	runTrack(ctx, args[1:])
	case "playlist":
		return	runPlaylist(ctx, args[1:])
	case "favorites":
		return	runFavorites(ctx, args[1:])
	case "album":
		return	runAlbum(ctx, args[1:])
	default:
		return fmt.Errorf("Unknown command %q", args[0])
	}
	//uid, _ := strconv.Atoi(os.Getenv("DEEZER_UID"))
	//ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	//defer stop()
	//downloader, err := InitDownloader(ctx, client, "downloads", false)
	//if err != nil{
	//	log.Fatal(err)
	//}
}

func runTrack(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("track", flag.ContinueOnError)
	outDir := fs.String("out", "downloads", "output directory")
	force := fs.Bool("force", false, "overwrite existing files")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: dzscrape track [flags] <track-id>")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		fs.Usage()
		return errors.New("track: expected exactly one track ID")
	}

	id, err := strconv.Atoi(fs.Arg(0))
	if err != nil || id <= 0 {
		return fmt.Errorf("Invalid track ID %q", fs.Arg(0))
	}

	client := InitClient(os.Getenv("DEEZER_ARL"))
	downloader, err := InitDownloader(ctx, client, *outDir, *force)
	
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

func runPlaylist(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("playlist", flag.ContinueOnError)
	force := fs.Bool("force", false, "overwrite existing files")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: dzscrape playlist [flags] <playlist-id>")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	id, err := strconv.Atoi(fs.Arg(0))
	if err != nil || id <= 0 {
		return fmt.Errorf("Invalid playlist ID %q", fs.Arg(0))
	}
	
	client := InitClient(os.Getenv("DEEZER_ARL"))
	playlist, err := client.GetPlaylist(ctx, id)

	outDir := fs.String("out", fmt.Sprintf("downloads/%s", playlist.Title), "output directory")


	if fs.NArg() != 1 {
		fs.Usage()
		return errors.New("Playlist: expected exactly one playlist ID")
	}

	downloader, err := InitDownloader(ctx, client, *outDir, *force)
	
	if err != nil {
		return err
	}
	if err := downloader.DownloadPlaylist(ctx, id); err != nil {
		return err
	}
	fmt.Printf("Playlist saved to %s\n", *outDir)
	return nil
}

func runFavorites(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("favorites", flag.ContinueOnError)
	outDir := fs.String("out", "downloads/favorites", "output directory")
	force := fs.Bool("force", false, "overwrite existing files")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: dzscrape favorites [flags] <user-id>")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		fs.Usage()
		return errors.New("favorites: expected exactly one user ID")
	}

	id, err := strconv.Atoi(fs.Arg(0))
	if err != nil || id <= 0 {
		return fmt.Errorf("Invalid user ID %q", fs.Arg(0))
	}

	client := InitClient(os.Getenv("DEEZER_ARL"))
	downloader, err := InitDownloader(ctx, client, *outDir, *force)
	
	if err != nil {
		return err
	}
	if err := downloader.DownloadFavorite(ctx, id); err != nil {
		return err
	}
	fmt.Printf("Saved: %s", *outDir)
	return nil
}

func runAlbum(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("album", flag.ContinueOnError)
	force := fs.Bool("force", false, "overwrite existing files")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: dzscrape album [flags] <album-id>")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	id, err := strconv.Atoi(fs.Arg(0))
	if err != nil || id <= 0 {
		return fmt.Errorf("Invalid album ID %q", fs.Arg(0))
	}
	
	client := InitClient(os.Getenv("DEEZER_ARL"))
	album, err := client.GetAlbum(ctx, id)

	outDir := fs.String("out", fmt.Sprintf("downloads/albums/%s", album.Title) , "output directory")


	if fs.NArg() != 1 {
		fs.Usage()
		return errors.New("Album: expected exactly one album ID")
	}

	downloader, err := InitDownloader(ctx, client, *outDir, *force)
	
	if err != nil {
		return err
	}
	if err := downloader.DownloadAlbum(ctx, id); err != nil {
		return err
	}
	fmt.Printf("Album saved to %s\n", *outDir)
	return nil
}
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
		os.Exit(1)
	}
	//albumID := 6651488
	//songID := 67929161 // track with a single artsit
	//songID := 67929169 // track with multiple artists
	//songID :=	"-3040362302" // track that was imported locally
	//_, err = downloader.DownloadTrack(ctx, songID)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//downloader.DownloadTrack(ctx, songID)

}
