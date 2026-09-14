package metadata

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/As71er/once/internal/utils"
)

func Extract(path string) (*Tags, error) {
	cmd := exec.Command("ffprobe", "-hide_banner", "-v", "error", "-print_format", "json", "-show_format", "-show_streams", path)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed at extracting tags: %w", err)
	}

	var ffprobeOut FFprobeAudioOutput
	err = json.Unmarshal(out, &ffprobeOut)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	// ffprobe keeps the case sensitive tags from Vorbis vs ID3
	rawTags := make(map[string]string)
	for k, v := range ffprobeOut.Format.Tags {
		rawTags[strings.ToLower(k)] = strings.TrimSpace(v)
	}

	value, _ := findTag(rawTags, "track", "trck", "tracknumber")
	before, _, _ := strings.Cut(value, "/")
	trackNumber := utils.ParseInt(before, 0)

	value, _ = findTag(rawTags, "disc", "discnumber")
	before, _, _ = strings.Cut(rawTags["disc"], "/")
	discNumber := utils.ParseInt(before, 0)

	totalTracks := utils.ParseInt(rawTags["totaltracks"], 0)
	totalDiscs := utils.ParseInt(rawTags["totaldiscs"], 0)

	tags := Tags{
		Title:       rawTags["title"],
		Artist:      rawTags["artist"],
		Album:       rawTags["album"],
		AlbumArtist: rawTags["album_artist"],
		Composer:    rawTags["composer"],
		Genre:       rawTags["genre"],
		Date:        rawTags["date"],
		Lyrics:      rawTags["lyrics"],
		TrackNumber: trackNumber,
		DiscNumber:  discNumber,
		TotalTracks: totalTracks,
		TotalDiscs:  totalDiscs,
		Streams:     ffprobeOut.Streams,
		Format: Format{
			FileName:   ffprobeOut.Format.FileName,
			FormatName: ffprobeOut.Format.FormatName,
			BitRate:    ffprobeOut.Format.BitRate,
			FileSize:   ffprobeOut.Format.FileSize,
		},
		RawTags: rawTags,
	}

	return &tags, nil
}

func HasCover(in string) (bool, error) {
	cmd := exec.Command("ffprobe", "-hide_banner", "-v", "error", "-print_format", "json", "-show_streams", in)
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed at extracting cover: %w", err)
	}

	var ffprobeOut FFprobeCoverOutput
	if err := json.Unmarshal(out, &ffprobeOut); err != nil {
		return false, fmt.Errorf("unmarshal failed: %w", err)
	}

	for _, stream := range ffprobeOut.Streams {
		if stream.CodecType == "video" &&
			stream.Disposition.AttachedPic == 1 {
			return true, nil
		}
	}

	return false, nil
}

func ExtractCover(in string, out io.Writer) error {
	cmd := exec.Command("ffmpeg", "-hide_banner", "-v", "error", "-i", in,
		"-map", "0:v:0?", "-frames:v", "1",
		"-c:v", "png", "-f", "image2pipe", "pipe:1")

	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func ExtractCoverThumbnail(in string, out io.Writer) error {
	cmd := exec.Command("ffmpeg", "-hide_banner", "-v", "error", "-i", in,
		"-map", "0:v:0?", "-vf", "scale='min(300,iw)':'min(300,ih)':force_original_aspect_ratio=decrease",
		"-frames:v", "1", "-q:v", "3", "-c:v", "mjpeg", "-f", "image2pipe", "pipe:1",
	)

	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func findTag(rawTags map[string]string, tags ...string) (string, string) {
	for _, tag := range tags {
		if value, ok := rawTags[tag]; ok {
			return value, tag
		}
	}
	return "", ""
}
