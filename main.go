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
	Id             int     `json:"id"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Duration       int     `json:"duration"`
	Public         bool    `json:"public"`
	Is_loved_track bool    `json:"is_loved_track"`
	Collaborative  bool    `json:"collaborative"`
	Nb_tracks      int     `json:"nb_tracks"`
	Fans           int     `json:"fans"`
	Link           string  `json:"link"`
	Share          string  `json:"share"`
	Picture        string  `json:"picture"`
	Picture_small  string  `json:"picture_small"`
	Picture_medium string  `json:"picture_medium"`
	Picture_big    string  `json:"picture_big"`
	Picture_xl     string  `json:"picture_xl"`
	Checksum       string  `json:"checksum"`
	Tracklist      string  `json:"tracklist"`
	Creation_date  string  `json:"creation_date"`
	Add_date       string  `json:"add_date"`
	Mod_date       string  `json:"mod_date"`
	Md5_image      string  `json:"md5_image"`
	Picture_type   string  `json:"picture_type"`
	Creator        Creator `json:"creator"`
	Tracks         Tracks  `json:"tracks"`
}
	
type Creator struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Tracklist string `json:"tracklist"`
	Type      string `json:"type"`
}

type Tracks struct {
	Data     []PlaylistData `json:"data"`
	Checksum string         `json:"checksum"`
}

type PlaylistData struct {
	Id                      int			`json:"id"`
	Readable                bool   		`json:"readable"`
	Title                   string 		`json:"title"`
	Title_short             string 		`json:"title_short"`
	Title_version           string 		`json:"title_version"`
	Isrc                    string 		`json:"isrc"`
	Link                    string 		`json:"link"`
	Duration                int    		`json:"duration"`
	Rank                    int    		`json:"rank"`
	Explicit_lyrics         bool   		`json:"explicit_lyrics"`
	Explicit_content_lyrics int    		`json:"explicit_content_lyrics"`
	Explicit_content_cover  int    		`json:"explicit_content_cover"`
	Preview                 string 		`json:"preview"`
	Md5_image               string 		`json:"md5_image"`
	Time_add                int			`json:"time_add"`

	Artist struct {
		Id        int    `json:"id"`
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
	Id				int    	`json:"id"`
	Title			string 	`json:"title"`
	Link			string	`json:"link"`
	Upc				string 	`json:"upc"`
	Cover			string 	`json:"cover"`
	Cover_small		string 	`json:"cover_small"`
	Cover_medium	string 	`json:"cover_medium"`
	Cover_big		string 	`json:"cover_big"`
	Cover_xl		string 	`json:"cover_xl"`
	Md5_image		string 	`json:"md5_image"`
	Genres	struct {
		Data	[]GenreData	`json:"data"`
	}`json:"genres"`
	Label			string	`json:"label"`
	Tracklist    	string 	`json:"tracklist"`
	Type         	string 	`json:"type"`
	Artist       	Artist 	`json:"artist"`
}

type GenreData	struct {
		Name	string	`json:"name"`
	}
type Artist struct {
	Id             	int    	`json:"id"`
	Name           	string 	`json:"name"`
	Link			string	`json:"link"` // new parameters
	Share			string	`json:"share"` // new parameters
	Picture			string 	`json:"picture"`
	Picture_small	string 	`json:"picture_small"`
	Picture_medium	string 	`json:"picture_medium"`
	Picture_big		string 	`json:"picture_big"`
	Picture_xl		string 	`json:"picture_xl"`
	Radio			bool	`json:"radio"` // new parameters
	Tracklist		string 	`json:"tracklist"`
	Type			string 	`json:"type"`
}
type FavoriteTracks struct {
	Id                      int    		`json:"id"`
	Readable                bool   		`json:"readable"`
	Title                   string 		`json:"title"`
	Link                    string 		`json:"link"`
	Duration                int    		`json:"duration"`
	Rank                    int    		`json:"rank"`
	Explicit_lyrics         bool   		`json:"explicit_lyrics"`
	Explicit_content_lyrics int    		`json:"explicit_content_lyrics"`
	Explicit_content_cover  int    		`json:"explicit_content_cover"`
	Md5_image               string 		`json:"md5_image"`
	Time_add                time.Time	`json:"time_add"`
	Album                   Album  		`json:"album"`
	Artist                  Artist 		`json:"artist"`
	Type                    string 		`json:"type"`
}

type User struct {
	Id             int    `json:"id"`
	Name           string `json:"name"`
	Link           string `json:"link"`
	Picture        string `json:"picture"`
	Picture_small  string `json:"picture_small"`
	Picture_medium string `json:"picture_medium"`
	Picture_big    string `json:"picture_big"`
	Picture_xl     string `json:"picture_xl"`
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
	Id						int				`json:"id"`
	Readable				bool			`json:"readable"`
	Title					string			`json:"title"`
	Title_short				string			`json:"title_short"`
	Title_version			string			`json:"title_version"`
	Isrc					string			`json:"isrc"`
	Link					string			`json:"link"`
	Share					string			`json:"share"`
	Duration				int				`json:"duration"`
	Track_position			int				`json:"track_position"`
	Disk_number				int				`json:"disk_number"`
	Rank					int				`json:"rank"`
	Release_date			string			`json:"release_date"`
	Explicit_lyrics			bool			`json:"explicit_lyrics"`
	Explicit_content_lyrics	int				`json:"explicit_content_lyrics"`
	Explicit_content_cover	int				`json:"explicit_content_cover"`
	Preview					string			`json:"preview"`
	Bpm						int				`json:"int"`
	Gain					float64			`json:"gain"`
	AvailableCountries		[]string		`json:"available_countries"`
	Contributors			[]Contributors	`json:"contributors"`
	Md5Image				string			`json:"md5_image"`
	TrackToken				string			`json:"track_token"`
	Artist					Artist			`json:"artist"`
	Album					Album			`json:"album"`

}

type Contributors struct {
	Id				int		`json:"id"`
	Name			string	`json:"name"`
	Link			string	`json:"link"`
	Share			string	`json:"share"`
	Picture			string	`json:"picture"`
	Picture_small	string	`json:"picture_small"`
	Picture_medium	string	`json:"picture_medium"`
	Picture_big		string	`json:"picture_big"`
	Picture_xl		string	`json:"picture_xl"`
	Radio			bool	`json:"radio"`
	Tracklist		string	`json:"tracklist"`
	Type			string	`json:"type"`
	Role			string	`json:"role"`
}

type GatewayResponse struct {
	DeezerError []DeezerError `json:"error"`
	Results	struct {
		User	struct {
			Options	struct {
				LicenseToken	string	`json:"license_token"`
			}	
		}	
	}		
}

var client *http.Client

func init_client() {
	jar, _ := cookiejar.New(nil)
	client = &http.Client{
		Jar: jar,
	}
}

func get_json(url string, v any) error {
	response, err := http.Get(url)

	if err != nil {
		return fmt.Errorf("get %s: %w", url, err)
	} else if response.StatusCode != http.StatusOK {
		defer response.Body.Close()
		return fmt.Errorf("get %s: unexpected status %s", url, response.Status)
	}

	err = json.NewDecoder(response.Body).Decode(v)
	if err != nil {
		return fmt.Errorf("decode %s: %w", url, err)
	}
	return nil
}

func fetch_playlist(id int) (FetchedDataPlaylist, error) {
	url := fmt.Sprintf("https://api.deezer.com/playlist/%d", id)
	var result Playlist
	err := get_json(url, &result)
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

func create_playlist_file(data FetchedDataPlaylist) error {
	re := regexp.MustCompile(`\s`)
	file_name := strings.ToLower(re.ReplaceAllString(data.PlaylistName, "_"))
	file_name = strings.ReplaceAll(file_name, "/", "_")

	music_list_file, err := os.Create(file_name + ".txt")
	if err != nil {
		return err
	}

	music_list := data.Tracks
	for _, track := range music_list {
		music_list_file.WriteString(track + "\n")
	}

	fmt.Printf("Tracklist has been written in %s.txt", file_name)
	return nil
}

func fetch_albums(uid int) ([]string, error) {
	url := fmt.Sprintf("https://api.deezer.com/user/%d/albums", uid)

	var result Albums
	get_json(url, &result)
	var albums []string
	for url != "" {
		var result Albums
		err := get_json(url, &result)
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

func get_username(uid int) (string, error) {
	url := fmt.Sprintf("https://api.deezer.com/user/%d", uid)
	var result User
	err := get_json(url, &result)
	if err != nil {
		return "", err
	}
	return result.Name, nil
}

func create_albums_file(data []string) {}

func fetch_favorite_tracks(uid int) ([]string, error) {
	url := fmt.Sprintf("https://api.deezer.com/user/%d/tracks", uid)
	var result Albums
	err := get_json(url, &result)
	if err != nil {
		return []string{}, err
	}

	var tracks []string
	for url != "" {
		var result Albums
		err := get_json(url, &result)
		if err != nil {
			return []string{}, err
		}

		url = result.Next
		for _, track := range result.Data {
			tracks = append(tracks, fmt.Sprintf("%s - %s", track.Artist.Name, track.Title))
		}
	}
	return tracks, nil
}

func calc_bf_key(track_id string) []byte {
	secret := []byte("g4el58wc0zvf9na1")
	hash :=	md5.Sum([]byte(track_id))
	hash_hex := hex.EncodeToString(hash[:])

	key := make([]byte, len(secret))
	copy(key, secret)
	
	for i := range len(hash) {
		key[i] = key[i] ^ hash_hex[i] ^ hash_hex[i + 16]
	}

	return key
}

func decrypt_chunk(data []byte, key []byte) ([]byte, error) {
	bf_iv := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}

	bf, err := blowfish.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypted := make([]byte, len(data))
	
	cipher.NewCBCDecrypter(bf, bf_iv).CryptBlocks(decrypted, data)
	return decrypted, nil
}

func decrypt_file(enc io.Reader, key []byte, dec io.Writer) error {  
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
        isEncrypted := i % 3 == 0
        isWholeBlock := n == blockSize
  
        if isEncrypted && isWholeBlock {
			data, err  = decrypt_chunk(data, key)
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

func get_license_token(arl string) (string, error) {
	ajax_url := "https://www.deezer.com/ajax/gw-light.php?method=deezer.getUserData&input=3&api_version=1.0&api_token="
	req, err := http.NewRequest("GET", ajax_url, http.NoBody)

	if err != nil {
		return "", err
	}

	req.Header.Set("Cookie", "arl=" + arl)
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

	license_token := result.Results.User.Options.LicenseToken
	return license_token, nil
}

func get_track_token(arl string, track_id string) (string, error) {
	api_endpoint := fmt.Sprintf("https://api.deezer.com/track/%s", track_id)
	req, err := http.NewRequest("GET", api_endpoint, http.NoBody)
	if err != nil {
		return "", nil
	}

	req.Header.Set("Cookie", "arl=" + arl + "; comeback=1")
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
	
	track_token := result.TrackToken
	return track_token, nil
}

func create_media_payload (license_token, track_token string) ([]byte, error) {
	type formats struct {
		Cipher string	`json:"cipher"`
		Format string	`json:"format"`
	}
	type media struct {
		Type	string		`json:"type"`
		Formats	[]formats	`json:"formats"`
	}

	type payload struct {
	LicenseToken	string `json:"license_token"`
	Media	[]media			`json:"media"`
	TrackTokens []string	`json:"track_tokens"`
	}


	pl := payload {
		LicenseToken: license_token,
		Media: []media{
			{
				Type: "FULL",
				Formats: []formats{
					{Cipher: "BF_CBC_STRIPE", Format: "MP3_128"},
				},
			},
		},
		TrackTokens: []string{track_token},
	}
	result, err := json.Marshal(pl)
	if err != nil {
		return []byte{}, err
	}
	return result, nil
}

func get_track_url(license_token, track_token string) (string, error) {
	type media struct {
		MediaType	string	`json:"media_type"`
		Cipher	struct {
			Type	string	`json:"type"`
		}	`json:"cipher"`
		Format		string	`json:"format"`
		Sources	 []struct {
			Url	string	`json:"url"`
		}	`json:"sources"`
	}

	type	data	struct {
		Media	[]media		`json:"media"`
	}

	type getUrlResponse struct {
		Data	[]data		`json:"data"`
	}

	get_url_endpoint := "https://media.deezer.com/v1/get_url"
	payload, err := create_media_payload(license_token, track_token)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", get_url_endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
		if err != nil {
			return "", err
		}

	defer resp.Body.Close()
	var result getUrlResponse
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
		if len(source.Url) == 0 {
			continue
		}
		url := source.Url
		return url, nil
			}
		}	
	}
	return "", err
}

func download_track(url string, track_info Track) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	file_name := fmt.Sprintf("%s - %s.mp3", track_info.Artist.Name, track_info.Title)
	file_name = strings.ReplaceAll(file_name, "/", "_")
	var enc bytes.Buffer

	_, err = io.Copy(&enc, resp.Body)
	
	track_id := fmt.Sprintf("%d", track_info.Id)
	key := calc_bf_key(track_id)
	if err != nil {
		return "", err
	}
	
	dec, err := os.Create(file_name)
	if err != nil {
		return "", err
	}

	decrypt_file(&enc, key, dec)

	return file_name, nil
}

func add_metadata(file_name string, track_info Track) error {
	var artists []string
	for _, contributor := range track_info.Contributors{
		artists = append(artists, contributor.Name)
	}

	album_url := fmt.Sprintf("https://api.deezer.com/album/%d", track_info.Album.Id)
	var album_info Album
	
	err := get_json(album_url, &album_info)
	if err != nil {
		return err
	}

	var genre string
	for _, g := range album_info.Genres.Data {
		genre = g.Name
	}

	const date_format = "2006-01-02"
	release_date, err := time.Parse(date_format, track_info.Release_date)
	err = taglib.WriteTags(file_name, map[string][]string{
		taglib.Artist:			artists,
		taglib.Title:			{track_info.Title},
		taglib.Album:			{track_info.Album.Title},
		taglib.AlbumArtist:		{track_info.Album.Artist.Name},
		taglib.TrackNumber:		{fmt.Sprintf("%d", track_info.Track_position)},
		taglib.ISRC:			{track_info.Isrc},
		taglib.BPM:				{fmt.Sprintf("%d", track_info.Bpm)},
		taglib.Genre:			{genre},
		taglib.Label:			{album_info.Label},
		"Year":					{fmt.Sprintf("%d", release_date.Year())},
		"Gain":					{fmt.Sprintf("%f", track_info.Gain)},
		}, 0)

	cover_url := track_info.Album.Cover_big

	resp, err := http.Get(cover_url)
	if err != nil {
		return err
	}

	var cover bytes.Buffer

	_, err = io.Copy(&cover, resp.Body)
	err = taglib.WriteImageOptions(
	file_name,
	cover.Bytes(), 0,
	"Front Cover",
	"Front artwork",
	"image/jpeg")

	if err != nil {
		return err
	}

	return nil
}

func get_track_info(track_id string) (Track, error) {
	api_endpoint := fmt.Sprintf("https://api.deezer.com/track/%s", track_id)
	var track Track
	err := get_json(api_endpoint, &track)
	if err !=nil {
		return Track{}, err
	}

	return track, nil
}


func main() {
	init_client()

	//user_id := os.Getenv("deezer_uid")
	song_id := "67929161" // track with a single artsit
	//song_id = "67929169" // track with multiple artists
	arl := os.Getenv("arl")

	lic_tk, err := get_license_token(arl)
	if err != nil{
		log.Fatal(err)
	}
	trk_tk, err := get_track_token(arl, song_id)
	if err != nil{
		log.Fatal(err)
	}

	trk_url, err := get_track_url(lic_tk, trk_tk)
	if err != nil{
		log.Fatal(err)
	}

	trk_info , err := get_track_info(song_id)
	if err != nil{
		log.Fatal(err)
	}

	file_name, err := download_track(trk_url, trk_info)
	if err != nil {
		log.Fatal(err)
	}

	err = add_metadata(file_name, trk_info)
	if err != nil {
		log.Fatal(err)
	}

	//data, err := fetch_playlist(playlist_id)
	//if err != nil {
	//	log.Fatal(err)
	//}
	
	////albums := fetch_albums(user_id)
	//for _, track := range data.Tracks {
	//	fmt.Println(track)time
	//}

	//favorite_tracks, err := fetch_favorite_tracks(user_id)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//for _, track := range favorite_tracks {
	//	fmt.Println(track)
	//}

}
