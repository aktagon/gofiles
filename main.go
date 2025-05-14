package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FileExplorerUI extends the base template for a file explorer application
type FileExplorerUI struct {
	app         *tview.Application
	grid        *tview.Grid
	header      *tview.TextView
	dirPane     *tview.Table
	contentPane *tview.TextView
	footer      *tview.TextView
	currentPath string
}

// NewFileExplorerUI creates and initializes a file explorer UI
func NewFileExplorerUI() *FileExplorerUI {
	ui := &FileExplorerUI{
		app:         tview.NewApplication(),
		grid:        tview.NewGrid(),
		header:      tview.NewTextView(),
		dirPane:     tview.NewTable(),
		contentPane: tview.NewTextView(),
		footer:      tview.NewTextView(),
	}

	// Get the current directory
	var err error
	ui.currentPath, err = os.Getwd()
	if err != nil {
		ui.currentPath = "."
	}

	ui.setupComponents()
	ui.setupLayout()
	ui.setupKeybindings()
	ui.loadDirectory(ui.currentPath)

	return ui
}

// setupComponents initializes individual UI components
func (ui *FileExplorerUI) setupComponents() {
	// Header setup
	ui.header.SetTextAlign(tview.AlignCenter)
	ui.header.SetDynamicColors(true)
	ui.header.SetText("[blue::b]File Explorer - " + ui.currentPath)
	ui.header.SetBackgroundColor(tcell.ColorDarkBlue)

	// Directory pane setup
	ui.dirPane.SetBorder(true)
	ui.dirPane.SetTitle("Directory Contents")
	ui.dirPane.SetBorderColor(tcell.ColorGreen)
	ui.dirPane.SetSelectable(true, false)
	ui.dirPane.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorDarkGreen).Foreground(tcell.ColorWhite))

	// Setup column headers
	ui.dirPane.SetCell(0, 0, tview.NewTableCell("Name").SetAttributes(tcell.AttrBold))
	ui.dirPane.SetCell(0, 1, tview.NewTableCell("Size").SetAttributes(tcell.AttrBold))
	ui.dirPane.SetCell(0, 2, tview.NewTableCell("Modified").SetAttributes(tcell.AttrBold))

	// Content view pane setup
	ui.contentPane.SetBorder(true)
	ui.contentPane.SetTitle("File Preview")
	ui.contentPane.SetBorderColor(tcell.ColorYellow)
	ui.contentPane.SetDynamicColors(true)
	ui.contentPane.SetWordWrap(true)
	ui.contentPane.SetText("Select a file to preview its contents")

	// Footer setup
	ui.footer.SetDynamicColors(true)
	ui.footer.SetText("[white]Keys: [yellow]↑/↓[white] Navigate | [yellow]Enter[white] Open | [yellow]Backspace[white] Go Up | [yellow]Ctrl-C[white] Quit")
	ui.footer.SetBackgroundColor(tcell.ColorDarkGray)
}

// setupLayout arranges UI components in a grid
func (ui *FileExplorerUI) setupLayout() {
	// Define grid layout
	ui.grid.SetRows(1, 0, 1)  // Header: 1 line, Main: flexible, Footer: 1 line
	ui.grid.SetColumns(0, 0)  // Two equal columns
	ui.grid.SetBorders(false) // No borders between cells

	// Add components to the grid
	ui.grid.AddItem(ui.header, 0, 0, 1, 2, 0, 0, false)      // Header spans both columns
	ui.grid.AddItem(ui.dirPane, 1, 0, 1, 1, 0, 0, true)      // Directory pane
	ui.grid.AddItem(ui.contentPane, 1, 1, 1, 1, 0, 0, false) // Content pane
	ui.grid.AddItem(ui.footer, 2, 0, 1, 2, 0, 0, false)      // Footer spans both columns

	// Set the grid as the root of the application
	ui.app.SetRoot(ui.grid, true)
}

// setupKeybindings configures application-wide keyboard shortcuts
func (ui *FileExplorerUI) setupKeybindings() {
	// Set up selection handler for the directory pane
	ui.dirPane.SetSelectionChangedFunc(func(row, column int) {
		if row > 0 { // Skip header row
			filename := ui.dirPane.GetCell(row, 0).Text
			fullPath := filepath.Join(ui.currentPath, filename)
			ui.previewFile(fullPath)
		}
	})

	// Set up selection handler for the directory pane
	ui.dirPane.SetSelectedFunc(func(row, column int) {
		if row > 0 { // Skip header row
			filename := ui.dirPane.GetCell(row, 0).Text

			if filename == ".." {
				// Go up one directory
				ui.currentPath = filepath.Dir(ui.currentPath)
				ui.loadDirectory(ui.currentPath)
				return
			}

			fullPath := filepath.Join(ui.currentPath, filename)
			fileInfo, err := os.Stat(fullPath)
			if err != nil {
				ui.setFooterError(err.Error())
				return
			}

			if fileInfo.IsDir() {
				// Navigate into the directory
				ui.currentPath = fullPath
				ui.loadDirectory(ui.currentPath)
			} else {
				// Preview the file
				ui.previewFile(fullPath)
			}
		}
	})

	// Set global keybindings
	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			ui.app.Stop()
			return nil
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			// Go up one directory
			ui.currentPath = filepath.Dir(ui.currentPath)
			ui.loadDirectory(ui.currentPath)
			return nil
		}
		return event
	})
}

// loadDirectory populates the directory pane with the contents of the given path
func (ui *FileExplorerUI) loadDirectory(path string) {
	// Clear the table
	ui.dirPane.Clear()

	// Re-add the header row
	ui.dirPane.SetCell(0, 0, tview.NewTableCell("Name").SetAttributes(tcell.AttrBold))
	ui.dirPane.SetCell(0, 1, tview.NewTableCell("Size").SetAttributes(tcell.AttrBold))
	ui.dirPane.SetCell(0, 2, tview.NewTableCell("Modified").SetAttributes(tcell.AttrBold))

	// Update header with current path
	ui.header.SetText("[blue::b]File Explorer - " + path)

	// Add parent directory entry
	ui.dirPane.SetCell(1, 0, tview.NewTableCell("..").SetTextColor(tcell.ColorBlue))
	ui.dirPane.SetCell(1, 1, tview.NewTableCell(""))
	ui.dirPane.SetCell(1, 2, tview.NewTableCell(""))

	// Read directory contents
	files, err := os.ReadDir(path)
	if err != nil {
		ui.setFooterError(err.Error())
		return
	}

	// Add files to the table
	row := 2
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}

		// Set the file name with appropriate color
		nameCell := tview.NewTableCell(file.Name())
		if file.IsDir() {
			nameCell.SetTextColor(tcell.ColorBlue)
		} else {
			nameCell.SetTextColor(tcell.ColorWhite)
		}
		ui.dirPane.SetCell(row, 0, nameCell)

		// Set the file size
		sizeText := "-"
		if !file.IsDir() {
			sizeText = formatSize(info.Size())
		}
		ui.dirPane.SetCell(row, 1, tview.NewTableCell(sizeText))

		// Set the modification time
		ui.dirPane.SetCell(row, 2, tview.NewTableCell(info.ModTime().Format("2006-01-02 15:04:05")))

		row++
	}

	// Select the first item (parent directory)
	ui.dirPane.Select(1, 0)
	ui.app.SetFocus(ui.dirPane)

	ui.setFooterStatus(path)
}

// previewFile shows a preview of the file in the content pane
func (ui *FileExplorerUI) previewFile(path string) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		ui.contentPane.SetText(fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	if fileInfo.IsDir() {
		ui.contentPane.SetText(fmt.Sprintf("Directory: %s\nContains %d items",
			path, countDirItems(path)))
		return
	}

	// Don't try to preview large files
	if fileInfo.Size() > 100*1024 { // 100KB limit
		ui.contentPane.SetText(fmt.Sprintf("File is too large to preview (%s)",
			formatSize(fileInfo.Size())))
		return
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		ui.contentPane.SetText(fmt.Sprintf("Error reading file: %s", err.Error()))
		return
	}

	// Check if it's a binary file
	if isBinary(content) {
		ui.contentPane.SetText(fmt.Sprintf("Binary file: %s\nSize: %s",
			path, formatSize(fileInfo.Size())))
		return
	}

	// Display the file content
	ui.contentPane.SetText(string(content))
}

// Helper function to set footer status
func (ui *FileExplorerUI) setFooterStatus(status string) {
	ui.footer.SetText(fmt.Sprintf("[white]%s | Keys: [yellow]↑/↓[white] Navigate | [yellow]Enter[white] Open | [yellow]Backspace[white] Go Up | [yellow]Ctrl-C[white] Quit", status))
}

// Helper function to set footer error
func (ui *FileExplorerUI) setFooterError(errMsg string) {
	ui.footer.SetText(fmt.Sprintf("[red]Error: %s[white] | Keys: [yellow]↑/↓[white] Navigate | [yellow]Enter[white] Open | [yellow]Backspace[white] Go Up | [yellow]Ctrl-C[white] Quit", errMsg))
}

// Start runs the application
func (ui *FileExplorerUI) Start() error {
	return ui.app.Run()
}

// Helper functions

// formatSize converts a file size in bytes to a human-readable string
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// countDirItems returns the number of items in a directory
func countDirItems(path string) int {
	files, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	return len(files)
}

// isBinary checks if data appears to be binary
// TODO: Make fast
func isBinary(data []byte) bool {
	// Check for null bytes or control characters, which suggest binary content
	for _, b := range data {
		if b == 0 || (b < 32 && b != '\n' && b != '\r' && b != '\t') {
			return true
		}
	}
	return false
}
