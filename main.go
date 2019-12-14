package main

import (
	"bufio"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dhowden/itl"
	"github.com/kennygrant/sanitize"
)

func main() {
	f, err := os.Open("D:\\Music\\iTunes\\iTunes Music Library.xml")
	check(err)

	library, err := itl.ReadFromXML(bufio.NewReader(f))
	check(err)

	CopyPlaylists(library, []string{
		// "Machinae Supremacy",
		"_Pop",
	})
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func CopyPlaylists(library itl.Library, playlistsToCopy []string) {
	tracksByPlaylist := make(map[string][]itl.Track)

	for _, playlist := range library.Playlists {
		playlistName := playlist.Name
		if SkipPlaylsit(playlistName) {
			continue
		}
		tracksByPlaylist[playlistName] = GetTracksForPlaylist(library, playlist)
		log.Println("Playlist:", playlistName)
	}

	for _, pl := range playlistsToCopy {
		tracks := tracksByPlaylist[pl]
		for _, track := range tracks {
			CopyTrackToDestination(track)
		}
	}
}

func SkipPlaylsit(playlist string) bool {
	excluded := []string{
		"Library",
		"Downloaded",
		"Music",
		"Movies",
		"TV Shows",
		"Podcasts",
		"Audiobooks",
		"Genius",
		"Top 25 Most Played",
		"Recently Played",
		"My Top Rated",
		"_Everything?",
	}
	for _, v := range excluded {
		if playlist == v {
			return true
		}
	}
	return false
}

func GetTracksForPlaylist(library itl.Library, playlist itl.Playlist) []itl.Track {
	var tracks []itl.Track
	for _, item := range playlist.PlaylistItems {
		track := library.Tracks[strconv.Itoa(item.TrackID)]
		tracks = append(tracks, track)
	}
	return tracks
}

func CopyTrackToDestination(track itl.Track) {
	src := TrackLocationToFilePath(track.Location)
	_, srcFile := filepath.Split(src)
	destRoot := filepath.Join("D:", "Mc", "files")

	var albumArtistOrAtrist string
	if track.AlbumArtist != "" {
		albumArtistOrAtrist = track.AlbumArtist
	} else {
		albumArtistOrAtrist = track.Artist
	}

	destDir := filepath.Join(
		destRoot,
		albumArtistOrAtrist+"-"+track.Album,
	)
	err := os.MkdirAll(destDir, os.ModeDir)
	check(err)
	dest := filepath.Join(destDir, srcFile)
	log.Println("Source:", src)
	log.Println("Destination:", dest)
	// CopyFile(src, dest)

	binary, lookErr := exec.LookPath("qaac64")
	check(lookErr)

	args := []string{src, "-v", "256", "-q", "2", "-o", dest}
	cmd := exec.Command(binary, args...)
	if err := cmd.Run(); err != nil {
		log.Println("Error:", dest, err)
		// check(err)
	}
	// log.Println("OK:", dest)
}

func TrackLocationToFilePath(location string) string {
	pathLike, err := url.QueryUnescape(location)
	check(err)
	path := strings.Replace(pathLike, "file://localhost/", "", 1)
	return sanitize.HTML(path)
}

func CopyFile(fromPath string, toPath string) {
	from, err := os.Open(fromPath)
	if err != nil {
		log.Fatal(err)
	}
	defer from.Close()

	to, err := os.OpenFile(toPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		log.Fatal(err)
	}
}
