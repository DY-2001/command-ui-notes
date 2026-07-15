package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	vaultDir string
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		os.Exit(1)
	}
	vaultDir = fmt.Sprintf("%s/.ui-notes", homeDir)
}

type item struct {
	title string
	desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct{
	newFileInput textinput.Model
	createFileInputVisible bool
	currentFile *os.File
	noteTextArea textarea.Model
	list list.Model
	showingList bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+q", "q":
			filterIsHappening := false
			if m.showingList && m.list.FilterState() == list.Filtering {
				filterIsHappening = true
			}
			if m.currentFile != nil || m.createFileInputVisible || filterIsHappening {
				return m, nil
			}
			return m, tea.Quit
		case "esc":
			if m.createFileInputVisible {
				m.newFileInput.SetValue("")
				m.createFileInputVisible = false
			}
			if m.currentFile != nil {
				m.currentFile.Close()
				m.currentFile = nil
				m.noteTextArea.SetValue("")
			}
			if m.showingList {
				if m.list.FilterState() == list.Filtering {
					m.list.ResetFilter()
				} else {
					m.showingList = false
				}
			}
			return m, nil
		case "ctrl+l":
			if m.currentFile != nil {
				return m, nil
			}
			if m.createFileInputVisible {
				m.createFileInputVisible = false
				m.newFileInput.SetValue("")
			}
			noteList := listFiles()
			m.list.SetItems(noteList)
			if m.showingList == true {
				if m.list.FilterState() == list.Filtering {
					m.list.ResetFilter()
				}
			}
			m.showingList = true
			return m, nil
		case "ctrl+n":
			if m.createFileInputVisible {
				return m, nil
			}
			if m.currentFile != nil {
				return m, nil
			}

			if m.showingList {
				if m.list.FilterState() == list.Filtering {
					m.list.ResetFilter()
				}
					
				m.showingList = false
			}
			m.createFileInputVisible = true
			m.newFileInput.SetValue("")
			return m, nil
		case "ctrl+s":
			if m.currentFile != nil {
				content := m.noteTextArea.Value()
				_, err := m.currentFile.WriteString(content)
				if err != nil {
					fmt.Println("Error saving file:", err)
				}
				m.currentFile.Close()
				m.currentFile = nil
				m.noteTextArea.SetValue("")
			}
		case "ctrl+d":
			if m.currentFile != nil {
				break
			}

			if m.showingList {
				if m.list.FilterState() == list.Filtering {
					break
				}
				item, ok := m.list.SelectedItem().(item)
				if ok {
					filePath := fmt.Sprintf("%s/%s", vaultDir, item.title)

					err := os.Remove(filePath)
					if err != nil {
						fmt.Println("Error deleting file:", err)
						return m, nil
					}

					m.list.SetItems(listFiles())

					// if len(m.list.Items()) == 0 {
					// 	m.showingList = false
					// }
				}

				return m, nil
			}
		case "enter":
			if m.currentFile != nil {
				break
			}

			if m.showingList {
				if m.list.FilterState() == list.Filtering {
					break
				}
				item, ok := m.list.SelectedItem().(item)
				if ok {
					filePath := fmt.Sprintf("%s/%s", vaultDir, item.title)
					content, err := os.ReadFile(filePath)
					if err != nil {
						fmt.Println("Error reading file:", err)
						return m, nil
					}
					m.noteTextArea.SetValue(string(content))
					f, err := os.OpenFile(filePath, os.O_RDWR, 0644)
					if err != nil {
						fmt.Println("Error opening file:", err)
						return m, nil
					}
					m.currentFile = f
					m.showingList = false
					m.noteTextArea.Focus()
				}
				return m, nil
			}

			fileName := m.newFileInput.Value()
			if fileName != "" {
				filePath := fmt.Sprintf("%s/%s.md", vaultDir, fileName)
				file, err := os.Create(filePath)
				if err != nil {
					fmt.Println("Error creating file:", err)
					return m, nil
				}

				m.currentFile = file
				m.createFileInputVisible = false
				m.newFileInput.SetValue("")
			}

			if m.createFileInputVisible {
				m.createFileInputVisible = false
			}
			return m, nil
		}
	}

	if( m.createFileInputVisible) {
		m.newFileInput, cmd = m.newFileInput.Update(msg)
	}

	if m.currentFile != nil {
		m.noteTextArea, cmd = m.noteTextArea.Update(msg)
	}

	if m.showingList {
		m.list, cmd = m.list.Update(msg)
	}


	return m, cmd
}

func (m model) View() string {
	var style = lipgloss.NewStyle().
	                        Bold(true).
							// Foreground(lipgloss.Color("16")).
							// Background(lipgloss.Color("254")).
							Padding(1)

	welcome := style.Render("Welcome to the Notes app! 🤗")

	help := "Ctrl+N: new file | ctrl+L: list files | Esc: back | ctrl+d: delete file | ctrl+S: save file | ctrl+Q: quit"

	view := ""
	if m.createFileInputVisible {
		view = m.newFileInput.View()
	} 

	if(m.currentFile != nil) {
		view = m.noteTextArea.View()
	}

	if m.showingList {
		view = m.list.View()
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s", welcome, view, help)
}

func initialModel() model {
	err := os.MkdirAll(vaultDir, 0750)
	if err != nil {
		fmt.Println("Error creating vault directory:", err)
		os.Exit(1)
	}
	ti := textinput.New()
	ti.Placeholder = "Hello Dev! whats your in mind today? 🤔"
	ti.Focus()
	ti.CharLimit = 156
	ti.PromptStyle = cursorStyle
	ti.Cursor.Style = cursorStyle
	ti.TextStyle = cursorStyle

	ta := textarea.New()
	ta.Placeholder = "Start typing your note here... 🧠"
	ta.ShowLineNumbers = false
	ta.Focus()
	
	noteList := listFiles()
	finalList := list.New(noteList, list.NewDefaultDelegate(), 50, 20)
	finalList.Title = "All Notes! 📝"

	finalList.Styles.Title = lipgloss.NewStyle().
	                                   Bold(true).
									   Foreground(lipgloss.Color("254")).
									   Margin(2, 0, 0, 0)
	                                               

	return model{newFileInput: ti, createFileInputVisible: false, noteTextArea: ta, list: finalList }
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}


func listFiles() []list.Item {
	items := make([]list.Item, 0)
	entries, err := os.ReadDir(vaultDir)
	if err != nil {
		fmt.Println("Error reading vault directory:", err)
		return items
	}
	
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			info, err := entry.Info()
			if err != nil {
				fmt.Println("Error getting file info:", err)
				continue
			}
			modTime := info.ModTime().Format("2006-01-02 15:04")
			items = append(items, item{title: entry.Name(), desc: fmt.Sprintf("Modified: %s", modTime)})
		}
	}

	return items
}