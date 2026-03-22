package components

import (
	"fmt"
	"strings"
	"zipcode/src/tools"

	"github.com/charmbracelet/lipgloss"
)

var (
	diffAddedBg   = lipgloss.NewStyle().Background(lipgloss.Color("22")).Foreground(lipgloss.Color("10"))
	diffRemovedBg = lipgloss.NewStyle().Background(lipgloss.Color("52")).Foreground(lipgloss.Color("9"))
	diffUnchanged = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	diffHunkSep   = lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Italic(true)
	diffFileLabel = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	diffOpBadge   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(lipgloss.Color("236")).Padding(0, 1)
	diffGutter    = lipgloss.NewStyle().Foreground(lipgloss.Color("239")).Width(4)
	diffAddSign   = lipgloss.NewStyle().Background(lipgloss.Color("22")).Foreground(lipgloss.Color("2")).Width(2)
	diffRemSign   = lipgloss.NewStyle().Background(lipgloss.Color("52")).Foreground(lipgloss.Color("1")).Width(2)
	diffCtxSign   = lipgloss.NewStyle().Foreground(lipgloss.Color("239")).Width(2)
)

type FileChangeViewer struct {
	FileName   string
	ChangeType string
	Content    string
	Patches    []tools.ParsedDiff
	Visible    bool
}

func CreateFileChangeViewer(fileName, changeType, content string, patches []tools.ParsedDiff) FileChangeViewer {
	return FileChangeViewer{
		FileName:   fileName,
		ChangeType: changeType,
		Content:    content,
		Patches:    patches,
		Visible:    true,
	}
}

func renderDiffLine(lineNum int, kind tools.DiffLineKind, content string) string {
	switch kind {
	case tools.DiffLineAdded:
		return diffAddSign.Render("+") + diffAddedBg.Render(" "+content)
	case tools.DiffLineRemoved:
		return diffRemSign.Render("-") + diffRemovedBg.Render(" "+content)
	default:
		return diffCtxSign.Render(" ") + diffUnchanged.Render(" "+content)
	}
}

func (f FileChangeViewer) View() string {
	if !f.Visible {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(diffFileLabel.Render(f.FileName) + "  " + diffOpBadge.Render(f.ChangeType) + "\n")
	sb.WriteString(diffUnchanged.Render(strings.Repeat("─", 60)) + "\n")

	switch f.ChangeType {
	case "patch":
		for _, diff := range f.Patches {
			for _, hunk := range diff.Hunks {
				hunkHeader := fmt.Sprintf("  @@ -%d,%d +%d,%d @@", hunk.OldStart, hunk.OldCount, hunk.NewStart, hunk.NewCount)
				sb.WriteString(diffHunkSep.Render(hunkHeader) + "\n")

				oldLine := hunk.OldStart
				newLine := hunk.NewStart
				for _, line := range hunk.Lines {
					switch line.Kind {
					case tools.DiffLineAdded:
						sb.WriteString(renderDiffLine(newLine, line.Kind, line.Content) + "\n")
						newLine++
					case tools.DiffLineRemoved:
						sb.WriteString(renderDiffLine(oldLine, line.Kind, line.Content) + "\n")
						oldLine++
					default:
						sb.WriteString(renderDiffLine(newLine, line.Kind, line.Content) + "\n")
						oldLine++
						newLine++
					}
				}
			}
		}

	case "create", "append", "replace":
		lines := strings.Split(f.Content, "\n")
		limit := len(lines)
		if limit > 30 {
			limit = 30
		}
		for i, line := range lines[:limit] {
			sb.WriteString(renderDiffLine(i+1, tools.DiffLineAdded, line) + "\n")
		}
		if len(lines) > 30 {
			sb.WriteString(diffUnchanged.Render(fmt.Sprintf("  ... %d more lines", len(lines)-30)) + "\n")
		}
	}

	sb.WriteString(diffUnchanged.Render(strings.Repeat("─", 60)) + "\n")
	return sb.String()
}
