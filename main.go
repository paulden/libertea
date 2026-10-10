package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	layoutName := flag.String("layout", envOrDefault(layoutEnvVar, defaultLayout), fmt.Sprintf(
		"keyboard layout, one of: %s (can also be set with %s)",
		strings.Join(LayoutNames(), ", "), layoutEnvVar,
	))
	colorMode := flag.String("color", envOrDefault(colorEnvVar, defaultColorMode), fmt.Sprintf(
		"color mode, one of: %s (can also be set with %s)",
		strings.Join(ColorModeNames(), ", "), colorEnvVar,
	))
	stratagemsPath := flag.String("stratagems", os.Getenv(stratagemsEnvVar), fmt.Sprintf(
		"YAML file with the stratagems to train on, defaults to the embedded list (can also be set with %s)",
		stratagemsEnvVar,
	))
	iconsMode := flag.String("icons", envOrDefault(iconsEnvVar, defaultIconsMode), fmt.Sprintf(
		"stratagem icons, one of: %s; they require a terminal supporting the kitty graphics protocol (can also be set with %s)",
		strings.Join(IconsModeNames(), ", "), iconsEnvVar,
	))
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(Version())
		return
	}

	layout, err := GetLayout(*layoutName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	colorProfile, err := ResolveColorProfile(*colorMode, os.Getenv, StdoutIsTTY())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	stratagems, err := LoadStratagems(*stratagemsPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	showIcons, err := ResolveIconsMode(*iconsMode, StdoutIsTTY())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	options := []tea.ProgramOption{tea.WithAltScreen()}
	var iconIDs map[string]int
	if showIcons {
		// Terminals store images per screen, so icons are transmitted once the
		// alternate screen is active, and bubbletea renders inline into it.
		fmt.Print(enterAltScreen)
		if iconIDs, err = TransmitIcons(os.Stdout, stratagems); err != nil {
			fmt.Print(deleteImages + exitAltScreen)
			fmt.Fprintf(os.Stderr, "Cannot transmit icons: %v\n", err)
			os.Exit(1)
		}
		options = nil
	}

	styles := NewStyles(colorProfile, iconIDs)
	model := NewModel(styles, layout, stratagems)

	_, err = tea.NewProgram(model, options...).Run()
	if showIcons {
		fmt.Print(deleteImages + exitAltScreen)
	}
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
