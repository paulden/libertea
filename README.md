# Libertea

Are you an IT worker that wishes you could do more for Super-Earth?
Do you want to help spread managed democracy?
Then you've come to the right place with `libertea`, a TUI-based stratagem hero to perfect your
skills before you prove yourself in the battlefield and put an end to our autocratic enemies!

[![asciicast](https://asciinema.org/a/54VFKPpaTbzt1WTDmmmHDAEAO.svg)](https://asciinema.org/a/54VFKPpaTbzt1WTDmmmHDAEAO)

## Run it

### Using Docker

You can run it from Docker as long as you attach a TTY to be able to interact with it.
Since the TUI uses colors, you should enable colors in the `xterm` started.

```
docker run -it -e "TERM=xterm-256color" ghcr.io/paulden/libertea:main
```

### Using binary

Download the binary matching your OS and architecture in the [releases](https://github.com/paulden/libertea/releases).

For Linux:
```
VERSION=$(curl https://api.github.com/repos/paulden/libertea/releases/latest | jq -r .tag_name)
curl -LO https://github.com/paulden/libertea/releases/download/$VERSION/libertea_Linux_x86_64.tar.gz
tar xvf libertea_Linux_x86_64.tar.gz libertea
./libertea # optionally, you can move it to a more friendly binary folder
```

### From source

You need to have Go >= 1.23 installed to run it from source

```
go get .
go run .
```

Or build it and run it.

```
go get .
go build .
./libertea
```

## Play

Type the arrow sequence of the displayed stratagem as fast as possible.
A wrong input blocks you for 2 seconds and you have to start the stratagem over.
Press `Esc` or `Ctrl+C` to quit.

### Keyboard layouts

Arrow keys always work. Letter keys depend on the selected layout:

| Layout   | Keys                                 |
|----------|--------------------------------------|
| `all`    | WASD, ZQSD and HJKL (default)        |
| `wasd`   | QWERTY keyboards                     |
| `zqsd`   | AZERTY keyboards                     |
| `vim`    | HJKL                                 |
| `arrows` | Arrow keys only                      |

Select it with the `-layout` flag or the `LIBERTEA_LAYOUT` environment variable (the flag wins):

```
./libertea -layout zqsd
LIBERTEA_LAYOUT=zqsd ./libertea
docker run -it -e "TERM=xterm-256color" -e "LIBERTEA_LAYOUT=zqsd" ghcr.io/paulden/libertea:main
```

### Colors

Colors are detected from the terminal (`TERM`, `COLORTERM`, `NO_COLOR`, `CLICOLOR_FORCE`).
Generic terminal names such as `xterm` (the default inside `docker run -t`) get 16 colors.
Without colors, a `^` cursor shows the next expected arrow.

Force a mode with the `-color` flag or the `LIBERTEA_COLOR` environment variable:
`auto` (default), `none`, `16`, `256`, `truecolor`.

```
./libertea -color 256
docker run -it -e "LIBERTEA_COLOR=256" ghcr.io/paulden/libertea:main
```

## Misc

- This is just a pet project to test [`bubbletea`](https://github.com/charmbracelet/bubbletea) and [lipgloss](https://github.com/charmbracelet/lipgloss) around the stratagem mechanism in Helldivers 2.
- The list of stratagems was fetched from Helldivers [wiki](https://helldivers.wiki.gg/wiki/Stratagems).

## TODO

- Improve styles
- Extract stratagem list to YAML
