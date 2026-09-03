package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const defaultVault = "/Users/uenokensuke/Documents/Obsidian Vault"

type dailySettings struct {
	Folder   string `json:"folder"`
	Format   string `json:"format"`
	Template string `json:"template"`
}

func main() {
	vault := flag.String("vault", defaultVault, "Obsidian vault path")
	dateArg := flag.String("date", "", "Target date as YYYY-MM-DD")
	flag.Parse()

	entry := strings.TrimSpace(strings.Join(flag.Args(), " "))
	if entry == "" {
		exitf("entry text must not be empty")
	}

	targetDate := time.Now()
	if *dateArg != "" {
		parsed, err := time.ParseInLocation("2006-01-02", *dateArg, time.Local)
		if err != nil {
			exitf("invalid --date %q: %v", *dateArg, err)
		}
		targetDate = parsed
	}

	settings, err := loadDailySettings(*vault)
	if err != nil {
		exitf("%v", err)
	}

	dailyPath := resolveDailyPath(*vault, targetDate, settings)
	if _, err := os.Stat(dailyPath); err != nil {
		if !os.IsNotExist(err) {
			exitf("stat daily note: %v", err)
		}
		if err := createDailyNote(dailyPath, *vault, targetDate, settings); err != nil {
			exitf("%v", err)
		}
	}

	currentBytes, err := os.ReadFile(dailyPath)
	if err != nil {
		exitf("read daily note: %v", err)
	}

	bullet := "- " + entry
	updated := appendToNippo(string(currentBytes), bullet)
	if err := os.WriteFile(dailyPath, []byte(updated), 0644); err != nil {
		exitf("write daily note: %v", err)
	}

	fmt.Printf("path: %s\n", dailyPath)
	fmt.Printf("bullet: %s\n", bullet)
}

func loadDailySettings(vault string) (dailySettings, error) {
	path := filepath.Join(vault, ".obsidian", "daily-notes.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return dailySettings{}, fmt.Errorf("read Obsidian daily notes settings %s: %w", path, err)
	}

	settings := dailySettings{
		Folder: "Dairy",
		Format: "YYYY/MM/YYYY-MM-DD",
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return dailySettings{}, fmt.Errorf("parse Obsidian daily notes settings %s: %w", path, err)
	}
	return settings, nil
}

func resolveDailyPath(vault string, date time.Time, settings dailySettings) string {
	format := settings.Format
	if format == "" {
		format = "YYYY-MM-DD"
	}
	relative := date.Format(momentToGoLayout(format)) + ".md"
	return filepath.Join(vault, settings.Folder, relative)
}

func createDailyNote(dailyPath, vault string, date time.Time, settings dailySettings) error {
	content := "### Nippo\n"
	if settings.Template != "" {
		templatePath := filepath.Join(vault, settings.Template)
		if filepath.Ext(templatePath) == "" {
			templatePath += ".md"
		}

		templateBytes, err := os.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("read configured daily note template %s: %w", templatePath, err)
		}
		content = renderTemplate(string(templateBytes), date, strings.TrimSuffix(filepath.Base(dailyPath), ".md"))
	}

	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if err := os.MkdirAll(filepath.Dir(dailyPath), 0755); err != nil {
		return fmt.Errorf("create daily note directory: %w", err)
	}
	if err := os.WriteFile(dailyPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("create daily note: %w", err)
	}
	return nil
}

func renderTemplate(template string, date time.Time, title string) string {
	replacements := map[string]string{
		"{{date}}":  date.Format("2006-01-02"),
		"{{time}}":  date.Format("15:04"),
		"{{title}}": title,
	}

	rendered := template
	for token, value := range replacements {
		rendered = strings.ReplaceAll(rendered, token, value)
	}
	return rendered
}

func appendToNippo(content, bullet string) string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	headingRE := regexp.MustCompile(`^###\s+Nippo\s*$`)
	nextSectionRE := regexp.MustCompile(`^#{1,3}\s+`)

	headingIndex := -1
	for i, line := range lines {
		if headingRE.MatchString(line) {
			headingIndex = i
			break
		}
	}

	if headingIndex == -1 {
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
			lines = append(lines, "")
		}
		lines = append(lines, "### Nippo", "", bullet)
		return strings.Join(lines, "\n") + "\n"
	}

	insertIndex := len(lines)
	for i := headingIndex + 1; i < len(lines); i++ {
		if nextSectionRE.MatchString(lines[i]) {
			insertIndex = i
			break
		}
	}

	for insertIndex > headingIndex+1 && strings.TrimSpace(lines[insertIndex-1]) == "" {
		insertIndex--
	}

	lines = append(lines[:insertIndex], append([]string{bullet}, lines[insertIndex:]...)...)
	return strings.Join(lines, "\n") + "\n"
}

func momentToGoLayout(format string) string {
	replacer := strings.NewReplacer(
		"YYYY", "2006",
		"YY", "06",
		"MM", "01",
		"DD", "02",
		"HH", "15",
		"mm", "04",
		"ss", "05",
	)
	return replacer.Replace(format)
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
