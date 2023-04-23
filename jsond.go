package main

type JSOND struct {
	Media Media `json:"media"`
}
type Media struct {
	Assets []Asset `json:"assets"`
}

type Asset struct {
	Type            string   `json:"type"`
	Slug            string   `json:"slug"`
	DisplayName     string   `json:"display_name"`
	Details         struct{} `json:"details"`
	Width           int      `json:"width"`
	Height          int      `json:"height"`
	Ext             string   `json:"ext"`
	Size            int      `json:"size"`
	Bitrate         int      `json:"bitrate"`
	Public          bool     `json:"public"`
	Status          int      `json:"status"`
	Progress        float64  `json:"progress"`
	Metadata        Metadata `json:"metadata"`
	URL             string   `json:"url"`
	CreatedAt       int      `json:"created_at"`
	SegmentDuration int      `json:"segment_duration,omitempty"`
	OptVbitrate     int      `json:"opt_vbitrate,omitempty"`
}

type Metadata struct {
	AvStreamMetadata string `json:"av_stream_metadata"`
	MaxBitrate       int    `json:"max_bitrate"`
	AverageBitrate   int    `json:"average_bitrate"`
	EarlyMaxBitrate  int    `json:"early_max_bitrate"`
}
