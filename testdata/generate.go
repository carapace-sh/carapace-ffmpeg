//go:generate echo "Generating test media files..."
//go:generate ffmpeg -hide_banner -y -f lavfi -i "color=c=black:s=2x2:d=2:r=1" -f lavfi -i "sine=frequency=440:duration=2:r=8000" -f lavfi -i "sine=frequency=880:duration=2:r=8000" -map 0:v -map 1:a -map 2:a -metadata:s:a:0 "language=eng" -metadata:s:a:1 "language=fre" -c:v libx264 -preset ultrafast -tune stillimage -crf 51 -c:a pcm_s16le multistream.mkv
//go:generate ffmpeg -hide_banner -y -f lavfi -i "color=c=black:s=2x2:d=2:r=1" -f lavfi -i "sine=frequency=440:duration=2:r=8000" -i subtitles.srt -map 0:v -map 1:a -map 2:s -c:v libx264 -preset ultrafast -crf 51 -c:a pcm_s16le -c:s srt -metadata:s:s:0 "language=eng" subtitles.mkv
//go:generate ffmpeg -hide_banner -y -f lavfi -i "color=c=black:s=2x2:d=2:r=1" -f lavfi -i "sine=frequency=440:duration=2:r=48000" -map 0:v -map 1:a -channel_layout 5.1 -c:v libx264 -preset ultrafast -crf 51 -c:a pcm_s16le surround.mkv
//go:generate ffmpeg -hide_banner -y -f lavfi -i "color=c=black:s=2x2:d=0.01:r=1" -f lavfi -i "sine=frequency=440:duration=0.01:r=8000" -map 0:v -map 1:a -c:v libx264 -preset ultrafast -crf 51 -pix_fmt yuv422p -c:a pcm_s16le pixfmt.mkv
//go:generate ffmpeg -hide_banner -y -f lavfi -i "sine=frequency=440:duration=0.01:r=8000" -c:a pcm_s16le audio_only.wav
//go:generate ffmpeg -hide_banner -y -f lavfi -i "color=c=black:s=2x2:d=0.01:r=1" -f lavfi -i "sine=frequency=440:duration=0.01:r=8000" -attach attachment.txt -metadata:s:t:0 "mimetype=text/plain" -map 0:v -map 1:a -c:v libx264 -preset ultrafast -crf 51 -c:a pcm_s16le attachment.mkv
//go:generate echo "Done generating test media files."

package testdata