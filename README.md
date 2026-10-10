# Libertea

Are you an IT worker that wishes you could do more for Super-Earth?
Do you want to help spread managed democracy?
Then you've come to the right place with `libertea`, a TUI-based stratagem hero to perfect your
skills before you prove yourself in the battlefield and put an end to our autocratic enemies!

[![CI](https://github.com/paulden/libertea/actions/workflows/ci.yml/badge.svg)](https://github.com/paulden/libertea/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/paulden/libertea?include_prereleases)](https://github.com/paulden/libertea/releases)

## Install

### Docker

Images are published for `linux/amd64` and `linux/arm64`. Attach a TTY with `-it`,
and forward your terminal type so that colors are detected:

```
docker run --rm -it -e TERM -e COLORTERM ghcr.io/paulden/libertea
```

| Tag | Content |
|-----|---------|
| `latest`, `main` | Latest commit on the `main` branch |
| `1`, `1.2`, `1.2.3` | Releases |
| `sha-<commit>` | A commit on the `main` branch, the 10 most recent are kept |

### Binary

Download the archive matching your OS and architecture from the [releases](https://github.com/paulden/libertea/releases).
On Linux and macOS:

```
OS=$(uname -s)    # Linux or Darwin
ARCH=$(uname -m)  # x86_64, aarch64 or arm64
[ "$ARCH" = aarch64 ] && ARCH=arm64
curl -fsSLO "https://github.com/paulden/libertea/releases/latest/download/libertea_${OS}_${ARCH}.tar.gz"
tar xzf "libertea_${OS}_${ARCH}.tar.gz" libertea
sudo install libertea /usr/local/bin/
```

Archives and images come with build provenance attestations. To check where they were built, with the [GitHub CLI](https://cli.github.com/):

```
gh attestation verify "libertea_${OS}_${ARCH}.tar.gz" --repo paulden/libertea
gh attestation verify oci://ghcr.io/paulden/libertea:latest --repo paulden/libertea
```

### Go

With Go 1.26 or later:

```
go install github.com/paulden/libertea@latest
```

Or from a clone of the repository: `go run .`

## Play

Type the arrow sequence of the displayed stratagem as fast as possible.
A wrong input blocks you for 2 seconds and you have to start the stratagem over.
Press `Esc` or `Ctrl+C` to quit.

### Options

Every option can be set with a flag or an environment variable, the flag wins.

| Flag | Environment variable | Default | Description |
|------|----------------------|---------|-------------|
| `-layout` | `LIBERTEA_LAYOUT` | `all` | Letter keys, see [keyboard layouts](#keyboard-layouts) |
| `-color` | `LIBERTEA_COLOR` | `auto` | Color mode: `auto`, `none`, `16`, `256`, `truecolor`, see [colors](#colors) |
| `-stratagems` | `LIBERTEA_STRATAGEMS` | embedded list | YAML file with the stratagems to train on, see [stratagems](#stratagems) |
| `-icons` | `LIBERTEA_ICONS` | `auto` | Stratagem icons: `auto`, `kitty`, `none`, see [icons](#icons) |
| `-version` | | | Print the version and exit |

### Keyboard layouts

Arrow keys always work. Letter keys depend on the selected layout:

| Layout   | Keys                                 |
|----------|--------------------------------------|
| `all`    | WASD, ZQSD and HJKL (default)        |
| `wasd`   | QWERTY keyboards                     |
| `zqsd`   | AZERTY keyboards                     |
| `vim`    | HJKL                                 |
| `arrows` | Arrow keys only                      |

```
libertea -layout zqsd
docker run --rm -it -e TERM -e COLORTERM -e LIBERTEA_LAYOUT=zqsd ghcr.io/paulden/libertea
```

### Colors

Colors are detected from the terminal (`TERM`, `COLORTERM`, `NO_COLOR`, `CLICOLOR_FORCE`).
Generic terminal names such as `xterm` (the default inside `docker run -t`) get 16 colors.
Without colors, a `^` cursor shows the next expected arrow.

If colors look wrong, force a mode:

```
libertea -color 256
```

### Stratagems

The stratagems come from [`stratagems.yaml`](internal/stratagem/stratagems.yaml), embedded in the binary.
To train on your own selection, write a file with the same format and pass it with `-stratagems`:

```yaml
stratagems:
  - name: Orbital Precision Strike
    category: offensive # one of offensive, supply, defensive, mission
    type: Orbital       # optional
    code: [right, right, up]
    icon: orbital-precision-strike # optional, one of the embedded icons
```

```
libertea -stratagems my-loadout.yaml
```

To refresh the embedded list and icons from the [Helldivers Wiki](https://helldivers.wiki.gg/wiki/Stratagems)
(requires `rsvg-convert` from librsvg):

```
go run ./cmd/update-stratagems
```

### Icons

Stratagem icons are shown next to their name in terminals supporting the
[kitty graphics protocol](https://sw.kovidgoyal.net/kitty/graphics-protocol/) with Unicode placeholders,
such as [kitty](https://sw.kovidgoyal.net/kitty/) and [Ghostty](https://ghostty.org/).
Support is detected by querying the terminal, other terminals only show the category colors.

If the detection fails in a terminal that supports it, or to disable icons:

```
libertea -icons kitty
libertea -icons none
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the development setup, the CI, releases and dependency updates.

## Misc

- This is just a pet project to test [`bubbletea`](https://github.com/charmbracelet/bubbletea) and [lipgloss](https://github.com/charmbracelet/lipgloss) around the stratagem mechanism in Helldivers 2.
- The list of stratagems is generated from the Helldivers [wiki](https://helldivers.wiki.gg/wiki/Stratagems), whose content is licensed under [CC BY-NC-SA 4.0](https://creativecommons.org/licenses/by-nc-sa/4.0).
- Stratagem icons are hand traced from the game assets by [Dogo314](https://helldivers.wiki.gg/wiki/User:Dogo314) for the Helldivers Wiki.
  The original artwork belongs to Arrowhead Game Studios, see [the icons README](internal/stratagem/icons/README.md).

## TODO

- Improve styles
