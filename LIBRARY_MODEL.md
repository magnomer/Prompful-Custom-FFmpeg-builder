# Library and options model

The Library & Options page has two different kinds of checked items.

## Included FFmpeg components

Rows marked `included` are built-in FFmpeg components such as `libavcodec`, `libavformat`, `libavfilter`, and native codecs/formats. They are checked and locked because they are part of the normal FFmpeg source build. They do not add MSYS2 packages and do not add `--enable-lib...` flags.

## External libraries

Unchecked rows are external libraries. Selecting one can add:

- MSYS2 packages,
- FFmpeg `./configure` flags,
- license effects,
- review warnings.

External libraries are not hidden in manual flags. Common external libraries must appear as named rows with plain and technical explanations.

## Advanced flags

The manual flags text area is only an escape hatch for flags that are not yet represented by a named checkbox. Every manual flag is shown again in Review before backend confirmation.
