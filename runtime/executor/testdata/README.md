# Synthetic container fixture

`tiny.webm` is a one-second 16×16 black VP9 video generated locally from a color
source on 2026-09-06. It contains no user media. Reproduction with a maintainer's
existing FFmpeg:

```sh
ffmpeg -f lavfi -i color=c=black:s=16x16:r=1:d=1 -an -c:v libvpx-vp9 -threads 1 tiny.webm
```

Tests consume the committed file; FFmpeg is not a test or user runtime
dependency. Synthetic in-memory EBML fixtures separately cover unknown-length
stream containers. Container validation does not claim codec playback success.
