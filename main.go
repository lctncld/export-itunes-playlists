package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/kennygrant/sanitize"

	"github.com/dhowden/itl"
)

func main() {
	// alb := "Proof / no vain"
	// albFix := toSafeName(alb)
	// log.Panicln(albFix)

	// brokenFileName := "file://localhost/E:/Music/Library/!2020/SHIKI%20-%20SHIKI%20Remix%20Compilation%20Breath%20of%20Air/Disk2/10%20Angelic%20layer%20(%C3%A9%C2%A1%C3%A9%C3%9A%C3%A9%C3%88%C3%A9-%C3%A9%C2%A6%C3%A9+%C3%A9%C3%9F%C3%A9%C2%B1%20Remix).mp3"
	// src := trackLocationToFilePath(brokenFileName)
	// log.Println(src)

	f, err := os.Open("D:\\Music\\iTunes\\iTunes Music Library.xml")
	check(err)

	library, err := itl.ReadFromXML(bufio.NewReader(f))
	check(err)

	copyPlaylists(library, []string{
		"Machinae Supremacy",
		"_Pop",
		"200104",
	})
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func copyPlaylists(library itl.Library, playlistsToCopy []string) {
	tracksByPlaylist := make(map[string][]itl.Track)

	for _, playlist := range library.Playlists {
		playlistName := playlist.Name
		if skipPlaylsit(playlistName) {
			continue
		}
		tracksByPlaylist[playlistName] = getTracksForPlaylist(library, playlist)
		log.Println("Playlist:", playlistName)
	}

	// var wg sync.WaitGroup

	for _, pl := range playlistsToCopy {
		tracks := tracksByPlaylist[pl]

		// jobs := make(chan int, 16)
		// go worker(tracks)
		processed := make([]string, 0)

		for _, track := range tracks {
			// go func(tr itl.Track) {
			// wg.Add(1)
			entry := copyTrackToDestination(track)
			processed = append(processed, entry)
			// wg.Done()
			// }(track)
		}

		err := printLines("E:\\Mc\\"+pl+".m3u", processed)
		check(err)
	}
	// wg.Wait()
}

func printLines(filePath string, values []string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, value := range values {
		fmt.Fprintln(f, value) // print values to f, one per line
	}
	return nil
}

// func worker(jobs <-chan int) {
// 	for n := range jobs {

// 	}
// }

func skipPlaylsit(playlist string) bool {
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

func getTracksForPlaylist(library itl.Library, playlist itl.Playlist) []itl.Track {
	var tracks []itl.Track
	for _, item := range playlist.PlaylistItems {
		track := library.Tracks[strconv.Itoa(item.TrackID)]
		tracks = append(tracks, track)
	}
	return tracks
}

var extensionByTrackKind = map[string]string{
	"MPEG audio file":           "mp3", // fuck this code auto format dude
	"Apple Lossless audio file": "m4a",
	"AAC audio file":            "m4a",
	"Purchased AAC audio file":  "m4a",
}

func copyTrackToDestination(track itl.Track) string {
	src := trackLocationToFilePath(track.Location)
	log.Println("Source:", src)
	// _, srcFile := filepath.Split(src)
	destRoot := filepath.Join("E:", "Mc", "files")

	var albumArtistOrAtrist string
	if track.AlbumArtist != "" {
		albumArtistOrAtrist = track.AlbumArtist
	} else {
		albumArtistOrAtrist = track.Artist
	}

	destDir := filepath.Join(
		destRoot,
		toSafeName(albumArtistOrAtrist)+"-"+toSafeName(track.Album),
	)

	log.Println("Mkdir", destDir)
	err := os.MkdirAll(destDir, os.ModeDir)
	check(err)

	dest := filepath.Join(destDir, destFileName(track))
	// log.Println("Destination:", dest)

	if track.Kind == "Apple Lossless audio file" {
		libmp3lame(src, dest)
	} else {
		copyFile(src, dest)
	}

	m3uPath := filepath.Join(
		destDir,
		destFileName(track),
	)
	return m3uPath
}

func destFileName(track itl.Track) string {
	trackNumber := fmt.Sprintf("%02d", track.TrackNumber)
	var extension string
	if track.Kind == "Apple Lossless audio file" {
		extension = "mp3"
	} else {
		extension = extensionByTrackKind[track.Kind]
	}
	if extension == "" {
		log.Panicln("Don't know track type", track.Kind)
	}
	destFileName := trackNumber + "-" + track.Name + "." + extension
	return toSafeName(destFileName)
}

var (
	separators = regexp.MustCompile(`[&_=+:]`)

	dashes = regexp.MustCompile(`[\-]+`)
)

func toSafeName(s string) string {

	// Start with lowercase string
	// fileName := strings.ToLower(s)
	// fileName = path.Clean(path.Base(fileName))

	// Remove any trailing space to avoid ending on -
	fileName := strings.Trim(s, " ")

	// Flatten accents first so that if we remove non-ascii we still get a legible name
	fileName = sanitize.Accents(fileName)

	// Replace certain joining characters with a dash
	fileName = separators.ReplaceAllString(fileName, "-")

	// Replace slashes
	fileName = strings.Replace(fileName, "/", "-", -1)
	fileName = strings.Replace(fileName, "\\", "-", -1)
	fileName = strings.Replace(fileName, "*", "-", -1)
	fileName = strings.Replace(fileName, "\"", "-", -1)
	fileName = strings.Replace(fileName, "?", "", -1)

	// Remove any multiple dashes caused by replacements above
	fileName = dashes.ReplaceAllString(fileName, "-")

	// NB this may be of length 0, caller must check
	return fileName

}

func trackLocationToFilePath(location string) string {
	pathLike, err := url.PathUnescape(location)
	check(err)
	path := strings.Replace(pathLike, "file://localhost/", "", 1)
	out := replace(path)
	return out
}

func replace(in string) string {
	var out string
	out = strings.Replace(in, "&#38;", "&", -1)
	return out
}

func qaac(src, dest string) {
	log.Println("Transcoding", src, dest)

	binary, lookErr := exec.LookPath("qaac64")
	check(lookErr)

	args := []string{"-i", src, "-v256", "-q2", "-o", dest}
	cmd := exec.Command(binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Println("Error:", dest, err)
		// check(err)
	}
	// log.Println("OK:", dest)
}

func libmp3lame(src, dest string) {
	log.Println("Transcoding", src, dest)

	binary, lookErr := exec.LookPath("ffmpeg")
	check(lookErr)

	args := []string{"-i", src, "-codec:a", "libmp3lame", "-q:a", "0", dest}
	cmd := exec.Command(binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Println("Error:", dest, err)
		// check(err)
	}
	// log.Println("OK:", dest)
}
func copyFile(fromPath string, toPath string) {
	log.Println("Copying", fromPath, toPath)
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
