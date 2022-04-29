package box

type Profile struct {
	AudioProfile      AudioProfile
	FileGlobalProfile FileGlobalProfile
	VideoProfile      VideoProfile
}

type AudioProfile struct {
}

type FileGlobalProfile struct {
}

type VideoProfile struct {
	VideoAvgBitrate  string
	PixelAspectRatio string
}
