package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/paulden/libertea/internal/buildinfo"
	"github.com/paulden/libertea/internal/keys"
	"github.com/paulden/libertea/internal/stratagem"
	"github.com/paulden/libertea/internal/terminal"
	"github.com/paulden/libertea/internal/ui"
)

// Set by GoReleaser and the Dockerfile with -ldflags "-X main.version=...".
var (
	version = "dev"
	commit  = ""
	date    = ""
)

// Every flag can also be set with an environment variable, the flag wins.
const (
	layoutEnvVar     = "LIBERTEA_LAYOUT"
	colorEnvVar      = "LIBERTEA_COLOR"
	stratagemsEnvVar = "LIBERTEA_STRATAGEMS"
	iconsEnvVar      = "LIBERTEA_ICONS"
)

func main() {
	layoutName := flag.String("layout", envOrDefault(layoutEnvVar, keys.Default), fmt.Sprintf(
		"keyboard layout, one of: %s (can also be set with %s)",
		strings.Join(keys.Names(), ", "), layoutEnvVar,
	))
	colorMode := flag.String("color", envOrDefault(colorEnvVar, terminal.DefaultColorMode), fmt.Sprintf(
		"color mode, one of: %s (can also be set with %s)",
		strings.Join(terminal.ColorModeNames(), ", "), colorEnvVar,
	))
	stratagemsPath := flag.String("stratagems", os.Getenv(stratagemsEnvVar), fmt.Sprintf(
		"YAML file with the stratagems to train on, defaults to the embedded list (can also be set with %s)",
		stratagemsEnvVar,
	))
	iconsMode := flag.String("icons", envOrDefault(iconsEnvVar, terminal.DefaultIconsMode), fmt.Sprintf(
		"stratagem icons, one of: %s; they require a terminal supporting the kitty graphics protocol (can also be set with %s)",
		strings.Join(terminal.IconsModeNames(), ", "), iconsEnvVar,
	))
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(buildinfo.Version(version, commit, date))
		return
	}

	layout, err := keys.Get(*layoutName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	colorProfile, err := terminal.ResolveColorProfile(*colorMode, os.Getenv, terminal.StdoutIsTTY())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	stratagems, err := stratagem.Load(*stratagemsPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	showIcons, err := terminal.ResolveIconsMode(*iconsMode, terminal.StdoutIsTTY())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	options := []tea.ProgramOption{tea.WithAltScreen()}
	var iconIDs map[string]int
	if showIcons {
		// Terminals store images per screen, so icons are transmitted once the
		// alternate screen is active, and bubbletea renders inline into it.
		fmt.Print(terminal.EnterAltScreen)
		if iconIDs, err = ui.TransmitIcons(os.Stdout, stratagems); err != nil {
			fmt.Print(terminal.DeleteImages + terminal.ExitAltScreen)
			fmt.Fprintf(os.Stderr, "Cannot transmit icons: %v\n", err)
			os.Exit(1)
		}
		options = nil
	}

	styles := ui.NewStyles(colorProfile, iconIDs)
	model := ui.NewModel(styles, layout, stratagems)

	_, err = tea.NewProgram(model, options...).Run()
	if showIcons {
		fmt.Print(terminal.DeleteImages + terminal.ExitAltScreen)
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
