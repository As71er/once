package metadata

type Stream struct {
	CodecType     string `json:"codec_type"`
	CodecName     string `json:"codec_name"`
	Duration      string `json:"duration"`
	SampleFmt     string `json:"sample_fmt"`
	SampleRate    string `json:"sample_rate"`
	BitDepth      string `json:"bits_per_raw_sample"`
	Channels      uint8  `json:"channels"`
	ChannelLayout string `json:"channel_layout"`
}

type Tags struct {
	Title       string            `json:"title"`
	Artist      string            `json:"artist"`
	Album       string            `json:"album"`
	AlbumArtist string            `json:"album_artist"`
	Composer    string            `json:"composer"`
	Genre       string            `json:"genre"`
	Date        string            `json:"date"`
	Lyrics      string            `json:"lyrics"`
	TrackNumber int               `json:"tracknumber"`
	DiscNumber  int               `json:"discnumber"`
	TotalTracks int               `json:"total_tracks"`
	TotalDiscs  int               `json:"total_discs"`
	Format      Format            `json:"format_info"`
	Streams     []Stream          `json:"streams,omitempty"`
	RawTags     map[string]string `json:"raw_tags,omitempty"`
}

type Format struct {
	FileName   string `json:"filename"`
	FormatName string `json:"format_name"`
	BitRate    string `json:"bit_rate"`
	FileSize   string `json:"size"`
}

type FFprobeAudioOutput struct {
	Streams []Stream      `json:"streams"`
	Format  FFprobeFormat `json:"format"`
}

type FFprobeFormat struct {
	FileName   string            `json:"filename"`
	FormatName string            `json:"format_name"`
	BitRate    string            `json:"bit_rate"`
	FileSize   string            `json:"size"`
	Tags       map[string]string `json:"tags"`
}

type StreamVideo struct {
	CodecType   string `json:"codec_type"`
	Disposition struct {
		AttachedPic int `json:"attached_pic"`
	} `json:"disposition"`
}

type FFprobeCoverOutput struct {
	Streams []StreamVideo `json:"streams"`
}
