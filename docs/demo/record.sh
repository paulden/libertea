#!/bin/sh
# Records the README demos, see the Dockerfile next to this script.
# Each demo trains on a single stratagem, so that the keys typed below always
# match the asked code.
set -eu

OUT=${1:-/out}
WIDTH=900
HEIGHT=470

Xvfb "$DISPLAY" -screen 0 "${WIDTH}x${HEIGHT}x24" -nolisten tcp >/dev/null 2>&1 &
sleep 1

# start_terminal runs a command in a kitty window filling the virtual screen.
start_terminal() {
	kitty --config NONE \
		-o font_family="JetBrains Mono" -o font_size=13 \
		-o background=#12110d -o foreground=#d8d4c4 -o cursor=#d8d4c4 \
		-o window_padding_width=12 -o remember_window_size=no \
		-o initial_window_width="$WIDTH" -o initial_window_height="$HEIGHT" \
		-o hide_window_decorations=yes -o cursor_blink_interval=0 \
		"$@" >/dev/null 2>&1 &
	KITTY=$!
	sleep 3
}

# start_shell opens a shell with a short prompt, playing on the given list.
start_shell() {
	start_terminal env PS1='$ ' LIBERTEA_STRATAGEMS="/demo/$1" bash --norc --noprofile
}

start_recording() {
	ffmpeg -loglevel error -y -f x11grab -framerate 30 -video_size "${WIDTH}x${HEIGHT}" -i "$DISPLAY" \
		-c:v libx264 -preset ultrafast -pix_fmt yuv420p /tmp/recording.mp4 &
	FFMPEG=$!
	sleep 0.5
}

# stop_recording converts the recording into a GIF with an optimized palette.
stop_recording() {
	kill -INT "$FFMPEG"
	wait "$FFMPEG" || true
	kill "$KITTY"
	wait "$KITTY" 2>/dev/null || true
	ffmpeg -loglevel error -y -i /tmp/recording.mp4 \
		-vf "fps=12,split[a][b];[a]palettegen=stats_mode=diff[p];[b][p]paletteuse=dither=bayer:bayer_scale=4" \
		"$OUT/$1.gif"
	echo "Recorded $OUT/$1.gif"
}

keys() {
	xdotool key --delay 170 "$@"
}

# keys_at types keys with the given delay in milliseconds, to vary the times.
keys_at() {
	delay=$1
	shift
	xdotool key --delay "$delay" "$@"
}

run() {
	xdotool type --delay 45 "$1"
	sleep 0.3
	xdotool key Return
	sleep 1.5
}

quit() {
	sleep 1
	xdotool key Escape
	sleep 0.8
}

# Colors and icons: successes, a wrong input and the penalty.
start_terminal env LIBERTEA_STRATAGEMS=/demo/railcannon.yaml libertea
start_recording
sleep 1
keys_at 200 Right Up Down Down Right
sleep 0.8
keys_at 260 Right Up Down Down Right
sleep 0.8
keys_at 170 Right Up
sleep 0.3
keys Up
sleep 2.4
keys_at 120 Right Up Down Down Right
sleep 0.8
keys_at 150 Right Up Down Down Right
sleep 1.5
stop_recording demo

# Every keyboard layout, played with its own keys.
start_shell reinforce.yaml
start_recording
run "libertea -layout arrows"
keys Up Down Right Left Up
quit
run "libertea -layout wasd"
keys w s d a w
quit
run "libertea -layout zqsd"
keys z s d q z
quit
run "libertea -layout vim"
keys k j l h k
quit
run "libertea -layout all"
keys w j Right q k
quit
stop_recording layouts

# 16 colors, then no colors with the cursor under the next arrow.
start_shell eat-17.yaml
start_recording
run "libertea -color 16"
keys Down Down Left Up Right
quit
run "libertea -color none"
xdotool key --delay 500 Down Down Left Up Right
quit
stop_recording colors

# Without icons, in terminals that do not support the kitty graphics protocol.
start_shell gatling-sentry.yaml
start_recording
run "libertea -icons none"
keys_at 220 Down Up Right Left
sleep 0.8
keys Down Down
sleep 2.4
keys_at 140 Down Up Right Left
quit
stop_recording no-icons

# Files written by the container belong to the owner of the output directory.
chown "$(stat -c %u:%g "$OUT")" "$OUT"/*.gif
