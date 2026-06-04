fname='./pkg/tag/testdata/with_tags/sample.vbr.mp3'
fname='./pkg/tag/testdata/with_tags/sample.padded.mp3'
go run ./cmd/tag/ "$fname"
ffprobe -v quiet -print_format json -show_format -show_streams "$fname"
