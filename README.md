# dzscrape
## Features
- Download tracks, albums, playlists, user's favorite tracks
- Write tags to downloaded audio files:\
  Artist (main and featured)\
  Title\
  Album\
  Year of release\
  Genre\
  Publisher\
  Position in album\
  BPM\
  Gain\
  ISRC
- Embed cover art
- Search for tracks, playlists and albums
- Show information about tracks, playlists and albums
- Automatically structure the downloaded library directory hierarchy
## Installation
```
# clone repo, build a program and make it executable
git clone https://github.com/krypxn36/dzscrape.git
cd dzscrape
go build
chmod +x dzscrape

# to be able to use it outisde of a repo directory add it into directory from your $PATH, for example
cp dzscrape $HOME/.local/bin/
```
## Usage
```
# view available options
dzscrape help

# searching 
dzscrape search track <query>
dzscrape search album <query>
dzscrape search playlist <query>

# downloading
dzscrape download track [flags] <track-id>
dzscrape download album [flags] <album-id>
dzscrape download playlist [flags] <playlist-id>
dzscrape download favorites [flags] <user-id>

# view info
dzscrape show track <track-id>
dzscrape show album <album-id>
dzscrape show playlist <playlist-id>
```
Flags for download commands:
```
-out <path> (output directory)
-quality <type> (choose audio quality - low (mp3 128kbps), medium (mp3 256kbps), high (mp3 320kbps), lossless (flac))
-force (overwrite existing files)
```
