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

## Misc

- This is just a pet project to test [`bubbletea`](https://github.com/charmbracelet/bubbletea) and [lipgloss](https://github.com/charmbracelet/lipgloss) around the stratagem mechanism in Helldivers 2.
- The list of stratagems was fetched from Helldivers [wiki](https://helldivers.wiki.gg/wiki/Stratagems).

## TODO

- Fix artifats in arrows in demo
- Improve styles
- Extract stratagem list to YAML
