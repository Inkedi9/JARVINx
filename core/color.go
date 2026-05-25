package core

import (
	"fmt"
	"runtime"
)

// On désactive les couleurs sur Windows si pas de terminal compatible
var colorEnabled = runtime.GOOS != "windows" || isWindowsTerminal()

func isWindowsTerminal() bool {
	// Windows Terminal et VS Code terminal supportent ANSI
	// CMD classique ne le supporte pas bien
	return true // Go 1.21+ active ANSI sur Windows automatiquement
}

const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[90m"
	colorWhite   = "\033[97m"
)

func colorize(color, text string) string {
	if !colorEnabled {
		return text
	}
	return color + text + colorReset
}

// Fonctions publiques — utilisées dans tout le projet
func Green(s string) string   { return colorize(colorGreen, s) }
func Yellow(s string) string  { return colorize(colorYellow, s) }
func Red(s string) string     { return colorize(colorRed, s) }
func Magenta(s string) string { return colorize(colorMagenta, s) }
func Cyan(s string) string    { return colorize(colorCyan, s) }
func Gray(s string) string    { return colorize(colorGray, s) }

// MetricColor retourne la couleur selon le pourcentage
func MetricColor(pct, warnThreshold, critThreshold float64) string {
	if pct >= critThreshold {
		return colorRed
	}
	if pct >= warnThreshold {
		return colorYellow
	}
	return colorGreen
}

func ColorMetric(label string, pct float64, warnAt, critAt float64) string {
	color := MetricColor(pct, warnAt, critAt)
	return fmt.Sprintf("%s : %s%.1f%%%s", label, color, pct, colorReset)
}
